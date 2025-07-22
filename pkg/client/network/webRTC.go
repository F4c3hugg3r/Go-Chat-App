package network

import (
	"fmt"
	"log"

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

func (p *Peer) JoinSession(chatClient *ChatClient) error {

	api, codecSelector, err := InitWebRTCAPI()
	if err != nil {
		return err
	}

	peerConnection, err := createPeerConnection(api)
	if err != nil {
		return err
	}

	s, err := mediadevices.GetUserMedia(mediadevices.MediaStreamConstraints{
		Audio: func(c *mediadevices.MediaTrackConstraints) {},
		Codec: codecSelector,
	})
	if err != nil {
		return err
	}

	// alle Media Tracks durchiterieren
	for _, track := range s.GetTracks() {
		track.OnEnded(func(err error) {
			fmt.Printf("Track (ID: %s) ended with error: %v\n",
				track.ID(), err)
		})

		// Transceiver ermöglicht senden und empfangen
		_, err = peerConnection.AddTransceiverFromTrack(track,
			webrtc.RTPTransceiverInit{
				Direction: webrtc.RTPTransceiverDirectionSendrecv,
			},
		)
		if err != nil {
			return err
		}
	}

	// SDP offer erzeugen, senden und empfangen
	offer, err := peerConnection.CreateOffer(nil)
	if err != nil {
		return err
	}

	err = peerConnection.SetLocalDescription(offer)
	if err != nil {
		return err
	}

	msg := chatClient.CreateMessage("", t.OfferSignal, offer.SDP, "")
	_, err = chatClient.PostMessage(msg, t.SignalWebRTC)
	if err != nil {
		log.Printf("%v: error posting signal offer", err)
	}

	peerConnection.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate == nil {
			return
		}
		candidateMsg := chatClient.CreateMessage("", t.ICECandidate, candidate.ToJSON().Candidate, "")
		_, err := chatClient.PostMessage(candidateMsg, t.SignalWebRTC)
		if err != nil {
			log.Printf("%v: error posting ICE candidate", err)
		}
	})

	go pollSignals(peerConnection, p.SignalChan, chatClient, p.Ctx)

	// Handler für eingehende Tracks
	peerConnection.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		fmt.Printf("Received track: ID=%s, kind=%s\n", track.ID(), track.Kind())
		if track.Kind() != webrtc.RTPCodecTypeAudio {
			return
		}

		otoCtx, ready, err := oto.NewContext(&oto.NewContextOptions{
			SampleRate:   48000,
			ChannelCount: 1,
			Format:       oto.FormatSignedInt16LE,
			BufferSize:   0,
		})
		if err != nil {
			log.Println("oto.NewContext error:", err)
			return
		}
		<-ready

		decoder, err := opusDec.NewDecoder(48000, 1)
		if err != nil {
			log.Println("opus.NewDecoder error:", err)
			return
		}
		// (TODO) jitter buffer für stabilere Audioqualität
		// TODO player und context schließen, wenn verbindung beendet wird

		// RTP Pakete von Opus in PCM decodieren
		reader := t.NewOpusRTPReader(track, decoder)
		player := otoCtx.NewPlayer(reader)
		go player.Play()
	})

	return nil
}

func pollSignals(peerConnection *webrtc.PeerConnection, signalChan chan *t.Response, chatClient *ChatClient, ctx context.Context) {
	for {
		select {
		case rsp := <-signalChan:
			switch rsp.RspName {
			case "signal answer":
				go HandleIncomingAnswer(peerConnection, rsp.Content)
			case "signal candidate":
				go HandleIncomingCandidate(peerConnection, rsp.Content)
			case "signal offer":
				go HandleIncomingOffer(peerConnection, chatClient, rsp.Content)
			}
		case <-ctx.Done():
			return
		}

	}
}

func HandleIncomingOffer(peerConnection *webrtc.PeerConnection, chatClient *ChatClient, SDPOffer string) {
	err := peerConnection.SetRemoteDescription(webrtc.SessionDescription{
		Type: webrtc.SDPTypeOffer,
		SDP:  SDPOffer,
	})
	if err != nil {
		log.Printf("%v: error setting remote description", err)
	}

	// Erzeuge eine Answer
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		log.Printf("%v: error creating answer", err)
	}

	// Setze LocalDescription
	err = peerConnection.SetLocalDescription(answer)
	if err != nil {
		log.Printf("%v: error setting local description", err)

	}

	// Sende die Answer zurück
	msg := chatClient.CreateMessage("", t.AnswerSignal, answer.SDP, "")
	_, err = chatClient.PostMessage(msg, t.SignalWebRTC)
	if err != nil {
		log.Printf("%v: error posting signal answer", err)
	}
}

func HandleIncomingCandidate(peerConnection *webrtc.PeerConnection, ICECandidate string) {
	candidate := webrtc.ICECandidateInit{Candidate: ICECandidate}
	err := peerConnection.AddICECandidate(candidate)
	if err != nil {
		log.Printf("%v: error adding ice candidate", err)
	}
}

func HandleIncomingAnswer(peerConnection *webrtc.PeerConnection, SDPAnswer string) {
	err := peerConnection.SetRemoteDescription(webrtc.SessionDescription{
		Type: webrtc.SDPTypeAnswer,
		SDP:  SDPAnswer,
	})
	if err != nil {
		log.Printf("%v: error setting remote description", err)
	}
}

func InitWebRTCAPI() (*webrtc.API, *mediadevices.CodecSelector, error) {
	// opus audio codec konfiguration
	opusParams, err := opus.NewParams()
	if err != nil {
		return nil, nil, err
	}

	codeSelector := mediadevices.NewCodecSelector(mediadevices.WithAudioEncoders(&opusParams))
	// verwaltet codecs
	mediaEngine := webrtc.MediaEngine{}
	// trägt codecs ein
	codeSelector.Populate(&mediaEngine)
	// einheitliche API Instanz damit alle Peers die gleichen Codecs nutzen
	api := webrtc.NewAPI(webrtc.WithMediaEngine(&mediaEngine))

	return api, codeSelector, nil
}

func createPeerConnection(api *webrtc.API) (*webrtc.PeerConnection, error) {
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
		return nil, webrtc.ErrCertificateExpired
	}

	// Set the handler for ICE connection state
	// This will notify you when the peer has connected/disconnected
	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		fmt.Printf("ICE connection state has changed %s\n", connectionState.String())
	})

	return peerConnection, nil
}
