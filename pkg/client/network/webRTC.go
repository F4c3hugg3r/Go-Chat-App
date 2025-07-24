package network

import (
	"fmt"

	t "github.com/F4c3hugg3r/Go-Chat-Server/pkg/shared"
	"github.com/ebitengine/oto/v3"
	"github.com/pion/mediadevices"
	"github.com/pion/mediadevices/pkg/codec/opus"
	"github.com/pion/webrtc/v4"
	"golang.org/x/net/context"
	opusDec "gopkg.in/hraban/opus.v2"
)

type Peer struct {
	peerId     string
	SignalChan chan *t.Response
	Ctx        context.Context
	Cancel     context.CancelFunc
}

func NewPeer(clientId string) *Peer {
	peer := &Peer{SignalChan: make(chan *t.Response, 100)}
	// TODO cancel kontext wenn peer verbindung geschlossen wird
	peer.Ctx, peer.Cancel = context.WithCancel(context.Background())

	return peer
}

func (p *Peer) JoinSession(chatClient *ChatClient, logChannel chan t.Logg) error {
	const method = "JoinSession"

	api, codecSelector, err := InitWebRTCAPI(logChannel)
	if err != nil {
		logChannel <- t.Logg{Text: fmt.Sprintf("WebRTC: Fehler bei InitWebRTCAPI: %v", err)}
		return err
	}

	logChannel <- t.Logg{Text: "WebRTC: API und CodecSelector initialisiert"}

	peerConnection, err := createPeerConnection(api, logChannel)
	if err != nil {
		logChannel <- t.Logg{Text: fmt.Sprintf("WebRTC: Fehler bei createPeerConnection: %v", err)}
		return err
	}

	logChannel <- t.Logg{Text: "WebRTC: PeerConnection erstellt"}

	s, err := mediadevices.GetUserMedia(mediadevices.MediaStreamConstraints{
		Audio: func(c *mediadevices.MediaTrackConstraints) {},
		Codec: codecSelector,
	})
	if err != nil {
		logChannel <- t.Logg{Text: fmt.Sprintf("WebRTC: Fehler bei GetUserMedia: %v", err)}
		return err
	}

	logChannel <- t.Logg{Text: "WebRTC: MediaStream erhalten"}

	// alle Media Tracks durchiterieren
	for _, track := range s.GetTracks() {
		logChannel <- t.Logg{Text: fmt.Sprintf("WebRTC: Füge Track hinzu: ID=%s, Kind=%s", track.ID(), track.Kind())}
		track.OnEnded(func(err error) {
			logChannel <- t.Logg{Text: fmt.Sprintf("WebRTC: Track (ID: %s) beendet mit Fehler: %v", track.ID(), err)}
		})

		// Transceiver ermöglicht senden und empfangen
		_, err = peerConnection.AddTransceiverFromTrack(track,
			webrtc.RTPTransceiverInit{
				Direction: webrtc.RTPTransceiverDirectionSendrecv,
			},
		)
		if err != nil {
			logChannel <- t.Logg{Text: fmt.Sprintf("WebRTC: Fehler bei AddTransceiverFromTrack: %v", err)}
			return err
		}
	}

	// SDP offer erzeugen, senden und empfangen
	offer, err := peerConnection.CreateOffer(nil)
	if err != nil {
		logChannel <- t.Logg{Text: fmt.Sprintf("WebRTC: Fehler bei CreateOffer: %v", err)}
		return err
	}

	logChannel <- t.Logg{Text: "WebRTC: SDP Offer erstellt"}

	err = peerConnection.SetLocalDescription(offer)
	if err != nil {
		logChannel <- t.Logg{Text: fmt.Sprintf("WebRTC: Fehler bei SetLocalDescription: %v", err)}
		return err
	}

	logChannel <- t.Logg{Text: "WebRTC: LocalDescription gesetzt"}

	msg := chatClient.CreateMessage("", t.OfferSignal, offer.SDP, "")
	_, err = chatClient.PostMessage(msg, t.SignalWebRTC)
	if err != nil {
		logChannel <- t.Logg{Text: fmt.Sprintf("WebRTC: Fehler beim Senden des Signal-Offers: %v", err)}
		return fmt.Errorf("%v: error posting signal offer", err)
	}

	logChannel <- t.Logg{Text: "WebRTC: Signal Offer gesendet"}

	peerConnection.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate == nil {
			logChannel <- t.Logg{Text: "WebRTC: ICECandidate ist nil"}
			return
		}
		candidateMsg := chatClient.CreateMessage("", t.ICECandidate, candidate.ToJSON().Candidate, "")
		_, err := chatClient.PostMessage(candidateMsg, t.SignalWebRTC)
		if err != nil {
			logChannel <- t.Logg{Text: fmt.Sprintf("WebRTC: Fehler beim Senden des ICE Candidates: %v", err)}
		}
	})

	go pollSignals(peerConnection, p.SignalChan, chatClient, p.Ctx, logChannel)

	// Handler für eingehende Tracks
	peerConnection.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		logChannel <- t.Logg{Text: fmt.Sprintf("WebRTC: Empfange Track: ID=%s, Kind=%s", track.ID(), track.Kind())}
		if track.Kind() != webrtc.RTPCodecTypeAudio {
			logChannel <- t.Logg{Text: "WebRTC: Track ist kein Audio, ignoriere"}
			return
		}

		otoCtx, ready, err := oto.NewContext(&oto.NewContextOptions{
			SampleRate:   48000,
			ChannelCount: 1,
			Format:       oto.FormatSignedInt16LE,
			BufferSize:   0,
		})
		if err != nil {
			logChannel <- t.Logg{Text: fmt.Sprintf("WebRTC: oto.NewContext error: %v", err)}
			return
		}
		<-ready
		logChannel <- t.Logg{Text: fmt.Sprintf("WebRTC: oto.Context bereit %v", err)}

		decoder, err := opusDec.NewDecoder(48000, 1)
		if err != nil {
			logChannel <- t.Logg{Text: fmt.Sprintf("WebRTC: opus.NewDecoder error: %v", err)}
			return
		}
		// (TODO) jitter buffer für stabilere Audioqualität
		// TODO player und context schließen, wenn verbindung beendet wird
		logChannel <- t.Logg{Text: "WebRTC: Opus Decoder bereit"}

		// RTP Pakete von Opus in PCM decodieren
		reader := t.NewOpusRTPReader(track, decoder)
		player := otoCtx.NewPlayer(reader)
		go player.Play()
		logChannel <- t.Logg{Text: "WebRTC: Audio Player gestartet"}
	})

	return nil
}

func pollSignals(peerConnection *webrtc.PeerConnection, signalChan chan *t.Response, chatClient *ChatClient, ctx context.Context, logChan chan t.Logg) {
	logChan <- t.Logg{Text: "WebRTC: pollSignals gestartet"}
	for {
		select {
		case rsp := <-signalChan:
			logChan <- t.Logg{Text: fmt.Sprintf("WebRTC: pollSignals - empfangen: %s", rsp.RspName)}
			switch rsp.RspName {
			case t.AnswerSignal:
				logChan <- t.Logg{Text: "WebRTC: HandleIncomingAnswer wird ausgeführt"}
				go HandleIncomingAnswer(peerConnection, rsp.Content, logChan)
			case t.ICECandidate:
				logChan <- t.Logg{Text: "WebRTC: HandleIncomingCandidate wird ausgeführt"}
				go HandleIncomingCandidate(peerConnection, rsp.Content, logChan)
			case t.OfferSignal:
				logChan <- t.Logg{Text: "WebRTC: HandleIncomingOffer wird ausgeführt"}
				go HandleIncomingOffer(peerConnection, chatClient, rsp.Content, logChan)
			}
		case <-ctx.Done():
			logChan <- t.Logg{Text: "WebRTC: pollSignals - Kontext beendet"}
			return
		}
	}
}

func HandleIncomingOffer(peerConnection *webrtc.PeerConnection, chatClient *ChatClient, SDPOffer string, logChan chan t.Logg) {
	logChan <- t.Logg{Text: "WebRTC: HandleIncomingOffer gestartet"}

	err := peerConnection.SetRemoteDescription(webrtc.SessionDescription{
		Type: webrtc.SDPTypeOffer,
		SDP:  SDPOffer,
	})
	if err != nil {
		logChan <- t.Logg{Text: fmt.Sprintf("WebRTC: Fehler beim SetRemoteDescription (Offer): %v", err)}
	}
	logChan <- t.Logg{Text: "WebRTC: RemoteDescription (Offer) gesetzt"}

	// Erzeuge eine Answer
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		logChan <- t.Logg{Text: fmt.Sprintf("WebRTC: Fehler beim CreateAnswer: %v", err)}
	}

	logChan <- t.Logg{Text: "WebRTC: SDP Answer erstellt"}

	// Setze LocalDescription
	err = peerConnection.SetLocalDescription(answer)
	if err != nil {
		logChan <- t.Logg{Text: fmt.Sprintf("WebRTC: Fehler beim SetLocalDescription (Answer): %v", err)}

	}
	logChan <- t.Logg{Text: "WebRTC: LocalDescription (Answer) gesetzt"}

	// Sende die Answer zurück
	msg := chatClient.CreateMessage("", t.AnswerSignal, answer.SDP, "")
	_, err = chatClient.PostMessage(msg, t.SignalWebRTC)
	if err != nil {
		logChan <- t.Logg{Text: fmt.Sprintf("WebRTC: Fehler beim Senden der Signal-Answer: %v", err)}
	}
	logChan <- t.Logg{Text: "WebRTC: Signal Answer gesendet"}

}

func HandleIncomingCandidate(peerConnection *webrtc.PeerConnection, ICECandidate string, logChan chan t.Logg) {
	logChan <- t.Logg{Text: "WebRTC: HandleIncomingCandidate gestartet"}

	candidate := webrtc.ICECandidateInit{Candidate: ICECandidate}
	err := peerConnection.AddICECandidate(candidate)
	if err != nil {
		logChan <- t.Logg{Text: fmt.Sprintf("WebRTC: Fehler beim AddICECandidate: %v", err)}
	}
	logChan <- t.Logg{Text: "WebRTC: ICECandidate hinzugefügt"}
}

func HandleIncomingAnswer(peerConnection *webrtc.PeerConnection, SDPAnswer string, logChan chan t.Logg) {
	logChan <- t.Logg{Text: "WebRTC: HandleIncomingAnswer gestartet"}
	err := peerConnection.SetRemoteDescription(webrtc.SessionDescription{
		Type: webrtc.SDPTypeAnswer,
		SDP:  SDPAnswer,
	})
	if err != nil {
		logChan <- t.Logg{Text: fmt.Sprintf("WebRTC: Fehler beim SetRemoteDescription (Answer): %v", err)}
	}
	logChan <- t.Logg{Text: "WebRTC: RemoteDescription (Answer) gesetzt"}
}

func InitWebRTCAPI(logChan chan t.Logg) (*webrtc.API, *mediadevices.CodecSelector, error) {
	logChan <- t.Logg{Text: "WebRTC: InitWebRTCAPI gestartet"}
	// opus audio codec konfiguration
	opusParams, err := opus.NewParams()
	if err != nil {
		logChan <- t.Logg{Text: fmt.Sprintf("WebRTC: Fehler bei opus.NewParams: %v", err)}
		return nil, nil, err
	}

	codeSelector := mediadevices.NewCodecSelector(mediadevices.WithAudioEncoders(&opusParams))
	// verwaltet codecs
	mediaEngine := webrtc.MediaEngine{}
	// trägt codecs ein
	codeSelector.Populate(&mediaEngine)
	// einheitliche API Instanz damit alle Peers die gleichen Codecs nutzen
	api := webrtc.NewAPI(webrtc.WithMediaEngine(&mediaEngine))

	logChan <- t.Logg{Text: "WebRTC: InitWebRTCAPI abgeschlossen"}
	return api, codeSelector, nil
}

func createPeerConnection(api *webrtc.API, logChan chan t.Logg) (*webrtc.PeerConnection, error) {
	logChan <- t.Logg{Text: "WebRTC: createPeerConnection gestartet"}
	// Define ICE servers
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}

	// Neue Peer Connection
	peerConnection, err := api.NewPeerConnection(config)
	if err != nil {
		logChan <- t.Logg{Text: fmt.Sprintf("WebRTC: Fehler bei NewPeerConnection: %v", err)}
		return nil, webrtc.ErrCertificateExpired
	}

	// Set the handler for ICE connection state
	// This will notify you when the peer has connected/disconnected
	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		logChan <- t.Logg{Text: fmt.Sprintf("WebRTC: ICE connection state has changed %s", connectionState.String())}
	})

	logChan <- t.Logg{Text: "WebRTC: createPeerConnection abgeschlossen"}

	return peerConnection, nil
}
