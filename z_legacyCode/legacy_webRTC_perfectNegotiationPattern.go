package legacy

// package network

// import (
// 	"fmt"
// 	"sync"

// 	t "github.com/F4c3hugg3r/Go-Chat-Server/pkg/shared"
// 	"github.com/ebitengine/oto/v3"
// 	"github.com/pion/mediadevices"
// 	"github.com/pion/mediadevices/pkg/codec/opus"
// 	_ "github.com/pion/mediadevices/pkg/driver/microphone"
// 	"github.com/pion/webrtc/v4"
// 	"golang.org/x/net/context"
// 	opusDec "gopkg.in/hraban/opus.v2"
// )

// type legacy_Peer struct {
// 	peerId      string
// 	oppClientId string
// 	SignalChan  chan *t.Response
// 	Ctx         context.Context
// 	Cancel      context.CancelFunc
// 	chatClient  *Client
// 	logChannel  chan t.Logg

// 	PeerConn      *webrtc.PeerConnection
// 	OtoContext    *oto.Context
// 	Players       []*oto.Player
// 	mediaStream   mediadevices.MediaStream
// 	codecSelector *mediadevices.CodecSelector
// 	api           *webrtc.API

// 	Polite      bool
// 	makingOffer bool
// 	ignoreOffer bool
// 	rollback    bool
// 	mu          *sync.RWMutex
// }

// func legacy_NewPeer(opposingId string, ownId string, logChannel chan t.Logg, chatClient *Client) *legacy_Peer {
// 	peer := &legacy_Peer{
// 		SignalChan:  make(chan *t.Response, 100),
// 		peerId:      opposingId,
// 		Polite:      opposingId > ownId,
// 		rollback:    false,
// 		logChannel:  logChannel,
// 		chatClient:  chatClient,
// 		oppClientId: opposingId,
// 		mu:          &sync.RWMutex{},
// 	}

// 	peer.Ctx, peer.Cancel = context.WithCancel(context.Background())

// 	return peer
// }

// func (p *legacy_Peer) Close(cancelAll bool) {
// 	if p.Cancel != nil && cancelAll {
// 		p.Cancel()
// 	}
// 	if p.PeerConn != nil {
// 		p.PeerConn.Close()
// 	}
// 	if p.mediaStream != nil {
// 		for _, track := range p.mediaStream.GetTracks() {
// 			track.Close()
// 		}
// 	}
// 	for _, player := range p.Players {
// 		if player != nil {
// 			player.Close()
// 		}
// 	}
// 	if p.OtoContext != nil {
// 		p.OtoContext.Suspend()
// 	}
// }

// func (p *legacy_Peer) JoinSession() error {
// 	var err error
// 	p.api, p.codecSelector, err = legacy_InitWebRTCAPI(p.logChannel)
// 	if err != nil {
// 		p.logChannel <- t.Logg{Text: fmt.Sprintf("WebRTC: Fehler bei InitWebRTCAPI: %v", err)}
// 		return err
// 	}

// 	p.logChannel <- t.Logg{Text: "WebRTC: API und CodecSelector initialisiert"}

// 	p.PeerConn, err = legacy_createPeerConnection(p.api, p.logChannel)
// 	if err != nil {
// 		p.logChannel <- t.Logg{Text: fmt.Sprintf("WebRTC: Fehler bei createPeerConnection: %v", err)}
// 		return err
// 	}

// 	p.logChannel <- t.Logg{Text: "WebRTC: PeerConnection erstellt"}

// 	p.PeerConn.OnTrack(p.OnTrackHandler)
// 	p.PeerConn.OnNegotiationNeeded(p.OnNegotiationNeededHandler)
// 	p.PeerConn.OnICECandidate(p.OnICECandidateHandler)

// 	err = p.AddTracksToPeerConnection(p.logChannel)
// 	if err != nil {
// 		p.logChannel <- t.Logg{Text: fmt.Sprintf("WebRTC: Fehler beim Hinzufügen von Tracks: %v", err)}
// 	}

// 	if !p.rollback {
// 		go p.pollSignals(p.SignalChan, p.chatClient, p.Ctx, p.logChannel, p.oppClientId)
// 	}

// 	return nil
// }

// func (p *legacy_Peer) OnNegotiationNeededHandler() {
// 	p.mu.Lock()
// 	defer p.mu.Unlock()

// 	p.logChannel <- t.Logg{Text: fmt.Sprintf("WebRTC: OnNegotiationNeeded started, SignalingState: %s", p.PeerConn.SignalingState())}

// 	// wenn connected und Zustand nicht Stable keine negotiation anstoßen
// 	if p.PeerConn.SignalingState() != webrtc.SignalingStateStable && p.PeerConn.ICEConnectionState() == webrtc.ICEConnectionStateConnected {
// 		p.logChannel <- t.Logg{Text: "WebRTC: OnNegotiationNeeded ignoriert, da SignalingState nicht stable && ICE state connected"}
// 		return
// 	}

// 	// wenn polite und Zustand nicht connected keine negotiation anstoßen
// 	if p.PeerConn.ICEConnectionState() != webrtc.ICEConnectionStateConnected && p.Polite && p.rollback {
// 		p.logChannel <- t.Logg{Text: "WebRTC: OnNegotiationNeeded ignoriert, da p.Polite und not connected & rollback"}
// 		return
// 	}

// 	p.makingOffer = true
// 	defer func() { p.makingOffer = false }()
// 	offer, err := p.PeerConn.CreateOffer(nil)
// 	if err != nil {
// 		p.logChannel <- t.Logg{Text: fmt.Sprintf("WebRTC: Fehler bei CreateOffer: %v", err)}
// 		return
// 	}

// 	p.logChannel <- t.Logg{Text: "WebRTC: SDP Offer erstellt"}

// 	err = p.PeerConn.SetLocalDescription(offer)
// 	if err != nil {
// 		p.logChannel <- t.Logg{Text: fmt.Sprintf("WebRTC: Fehler bei SetLocalDescription: %v", err)}
// 		return
// 	}

// 	p.logChannel <- t.Logg{Text: "WebRTC: LocalDescription gesetzt"}

// 	msg := p.chatClient.CreateMessage(t.OfferSignalFlag, "/signal", offer.SDP, p.oppClientId)
// 	_, err = p.chatClient.PostMessage(msg, t.SignalWebRTC)
// 	if err != nil {
// 		p.logChannel <- t.Logg{Text: fmt.Sprintf("WebRTC: Fehler beim senden des Offers %v", err)}
// 		return
// 	}
// 	p.logChannel <- t.Logg{Text: "WebRTC: Offer gesendet"}

// }

// func (p *legacy_Peer) OnTrackHandler(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
// 	p.logChannel <- t.Logg{Text: fmt.Sprintf("WebRTC: Empfange Track: ID=%s, Kind=%s", track.ID(), track.Kind())}
// 	if track.Kind() != webrtc.RTPCodecTypeAudio {
// 		p.logChannel <- t.Logg{Text: "WebRTC: Track ist kein Audio, ignoriere"}
// 		return
// 	}

// 	otoCtx, ready, err := oto.NewContext(&oto.NewContextOptions{
// 		SampleRate:   48000,
// 		ChannelCount: 1,
// 		Format:       oto.FormatSignedInt16LE,
// 		BufferSize:   0,
// 	})
// 	if err != nil {
// 		p.logChannel <- t.Logg{Text: fmt.Sprintf("WebRTC: oto.NewContext error: %v", err)}
// 		return
// 	}

// 	p.OtoContext = otoCtx
// 	<-ready
// 	p.logChannel <- t.Logg{Text: fmt.Sprintf("WebRTC: oto.Context bereit %v", err)}

// 	decoder, err := opusDec.NewDecoder(48000, 1)
// 	if err != nil {
// 		p.logChannel <- t.Logg{Text: fmt.Sprintf("WebRTC: opus.NewDecoder error: %v", err)}
// 		return
// 	}
// 	// legacy TODO jitter buffer für stabilere Audioqualität
// 	p.logChannel <- t.Logg{Text: "WebRTC: Opus Decoder bereit"}

// 	// RTP Pakete von Opus in PCM decodieren
// 	reader := t.NewOpusRTPReader(track, decoder)
// 	player := otoCtx.NewPlayer(reader)
// 	p.Players = append(p.Players, player)
// 	go player.Play()
// 	p.logChannel <- t.Logg{Text: "WebRTC: Audio Player gestartet"}
// }

// func (p *legacy_Peer) OnICECandidateHandler(candidate *webrtc.ICECandidate) {
// 	if candidate == nil {
// 		p.logChannel <- t.Logg{Text: "WebRTC: ICECandidate ist nil"}
// 		return
// 	}

// 	candidateMsg := p.chatClient.CreateMessage(t.ICECandidateFlag, "/signal", candidate.ToJSON().Candidate, p.oppClientId)
// 	_, err := p.chatClient.PostMessage(candidateMsg, t.SignalWebRTC)
// 	if err != nil {
// 		p.logChannel <- t.Logg{Text: fmt.Sprintf("WebRTC: Fehler beim Senden des ICE Candidates: %v", err)}
// 	}
// }

// func (p *legacy_Peer) AddTracksToPeerConnection(logChannel chan t.Logg) error {
// 	var err error
// 	p.mediaStream, err = mediadevices.GetUserMedia(mediadevices.MediaStreamConstraints{
// 		Audio: func(c *mediadevices.MediaTrackConstraints) {},
// 		Codec: p.codecSelector,
// 	})
// 	if err != nil {
// 		logChannel <- t.Logg{Text: fmt.Sprintf("WebRTC: Fehler bei GetUserMedia: %v", err)}
// 		return err
// 	}

// 	logChannel <- t.Logg{Text: "WebRTC: MediaStream erhalten"}

// 	for _, track := range p.mediaStream.GetTracks() {
// 		logChannel <- t.Logg{Text: fmt.Sprintf("WebRTC: Füge Track hinzu: ID=%s, Kind=%s", track.ID(), track.Kind())}
// 		track.OnEnded(func(err error) {
// 			logChannel <- t.Logg{Text: fmt.Sprintf("WebRTC: Track (ID: %s) beendet mit Fehler: %v", track.ID(), err)}
// 		})

// 		_, err = p.PeerConn.AddTransceiverFromTrack(track,
// 			webrtc.RTPTransceiverInit{
// 				Direction: webrtc.RTPTransceiverDirectionSendrecv,
// 			},
// 		)
// 		if err != nil {
// 			logChannel <- t.Logg{Text: fmt.Sprintf("WebRTC: Fehler bei AddTransceiverFromTrack: %v", err)}
// 			return err
// 		}
// 	}
// 	return nil
// }

// func (p *legacy_Peer) pollSignals(signalChan chan *t.Response, chatClient *Client, ctx context.Context, logChan chan t.Logg, oppClientId string) {
// 	logChan <- t.Logg{Text: "WebRTC: pollSignals gestartet"}
// 	for {
// 		select {
// 		case rsp := <-signalChan:
// 			logChan <- t.Logg{Text: fmt.Sprintf("WebRTC: pollSignals - empfangen: %s\n", rsp.RspName)}
// 			switch rsp.RspName {
// 			case t.ICECandidateFlag:
// 				logChan <- t.Logg{Text: "WebRTC: HandleIncomingCandidate wird ausgeführt"}
// 				p.HandleIncomingCandidate(rsp.Content, logChan)
// 			case t.OfferSignalFlag, t.AnswerSignalFlag:
// 				logChan <- t.Logg{Text: "WebRTC: HandleIncomingOffer wird ausgeführt"}
// 				p.HandleIncomingSignal(chatClient, rsp, logChan, oppClientId)
// 			}
// 		case <-ctx.Done():
// 			close(p.SignalChan)
// 			logChan <- t.Logg{Text: "WebRTC: pollSignals - Kontext beendet"}
// 			return
// 		}
// 	}
// }

// func (p *legacy_Peer) RestartPeerConnection(chatClient *Client, logChan chan t.Logg, oppClientId string) error {
// 	p.ignoreOffer = false
// 	p.Close(false)

// 	err := p.JoinSession()
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

// func (p *legacy_Peer) HandleIncomingSignal(chatClient *Client, rsp *t.Response, logChan chan t.Logg, oppClientId string) {
// 	p.mu.Lock()
// 	defer p.mu.Unlock()
// 	logChan <- t.Logg{Text: "WebRTC: HandleIncomingSignal gestartet"}

// 	readyForOffer := !p.makingOffer && p.PeerConn.SignalingState() == webrtc.SignalingStateStable
// 	logChan <- t.Logg{Text: fmt.Sprintf("WebRTC: readyForOffer: %t, !p.makingOffer: %t, signalingState: %s", readyForOffer, !p.makingOffer, p.PeerConn.SignalingState())}
// 	offerCollision := !readyForOffer && rsp.RspName == t.OfferSignalFlag
// 	logChan <- t.Logg{Text: fmt.Sprintf("WebRTC: offerCollision: %t, !readyForOffer: %t, rsp.Name: %s", offerCollision, !readyForOffer, rsp.RspName)}

// 	p.ignoreOffer = !p.Polite && offerCollision
// 	if p.ignoreOffer {
// 		logChan <- t.Logg{Text: "WebRTC: Impolite Peer ignoriert Offer wegen Kollision"}
// 		return
// 	}

// 	if offerCollision && p.Polite && p.PeerConn.SignalingState() != webrtc.SignalingStateStable {
// 		logChan <- t.Logg{Text: "WebRTC: Polite Peer führt Rollback durch"}

// 		var err error
// 		p.rollback = true
// 		defer func() { p.rollback = false }()
// 		err = p.RestartPeerConnection(chatClient, logChan, oppClientId)
// 		if err != nil {
// 			logChan <- t.Logg{Text: fmt.Sprintf("WebRTC: Fehler beim Rollback: %v", err)}
// 			return
// 		}
// 		logChan <- t.Logg{Text: "WebRTC: Rolled back PeerConnection"}
// 	}

// 	switch offerCollision {
// 	case false:
// 		logChan <- t.Logg{Text: "WebRTC: Keine Kollision"}
// 	case true:
// 		logChan <- t.Logg{Text: fmt.Sprintf("WebRTC: Offer Kollision, Peer polite: %t", p.Polite)}
// 	}

// 	var descriptionType webrtc.SDPType
// 	switch rsp.RspName {
// 	case t.AnswerSignalFlag:
// 		descriptionType = webrtc.SDPTypeAnswer
// 	case t.OfferSignalFlag:
// 		descriptionType = webrtc.SDPTypeOffer
// 	}

// 	err := p.PeerConn.SetRemoteDescription(webrtc.SessionDescription{
// 		Type: descriptionType,
// 		SDP:  rsp.Content,
// 	})
// 	if err != nil {
// 		logChan <- t.Logg{Text: fmt.Sprintf("WebRTC: Fehler beim SetRemoteDescription (%s): %v", descriptionType.String(), err)}
// 	}

// 	p.ignoreOffer = false
// 	logChan <- t.Logg{Text: fmt.Sprintf("WebRTC: RemoteDescription (%s) gesetzt", descriptionType.String())}

// 	if rsp.RspName == t.OfferSignalFlag {
// 		answer, err := p.PeerConn.CreateAnswer(nil)
// 		if err != nil {
// 			logChan <- t.Logg{Text: fmt.Sprintf("WebRTC: Fehler beim CreateAnswer: %v", err)}
// 			return
// 		}

// 		err = p.PeerConn.SetLocalDescription(answer)
// 		if err != nil {
// 			logChan <- t.Logg{Text: fmt.Sprintf("WebRTC: Fehler beim SetLocalDescription (Answer): %v", err)}
// 			return
// 		}

// 		logChan <- t.Logg{Text: "WebRTC: LocalDescription (Answer) gesetzt"}

// 		msg := chatClient.CreateMessage(t.AnswerSignalFlag, "/signal", answer.SDP, oppClientId)
// 		_, err = chatClient.PostMessage(msg, t.SignalWebRTC)
// 		if err != nil {
// 			logChan <- t.Logg{Text: fmt.Sprintf("WebRTC: Fehler beim Senden der Signal-Answer: %v", err)}
// 		}

// 		logChan <- t.Logg{Text: "WebRTC: Signal Answer gesendet"}
// 	}
// }

// func (p *legacy_Peer) HandleIncomingCandidate(ICECandidate string, logChan chan t.Logg) {
// 	p.mu.Lock()
// 	defer p.mu.Unlock()
// 	logChan <- t.Logg{Text: "WebRTC: HandleIncomingCandidate gestartet"}

// 	if p.ignoreOffer {
// 		logChan <- t.Logg{Text: "WebRTC: ICECandidate ignoriert wegen ignoreOffer"}
// 		return
// 	}

// 	// Prüfe, ob RemoteDescription gesetzt ist
// 	if p.PeerConn.RemoteDescription() == nil {
// 		logChan <- t.Logg{Text: "WebRTC: ICECandidate ignoriert - keine RemoteDescription"}
// 		return
// 	}

// 	candidate := webrtc.ICECandidateInit{Candidate: ICECandidate}

// 	err := p.PeerConn.AddICECandidate(candidate)
// 	if err != nil {
// 		logChan <- t.Logg{Text: fmt.Sprintf("WebRTC: Fehler beim AddICECandidate: %v", err)}
// 	}

// 	logChan <- t.Logg{Text: "WebRTC: ICECandidate hinzugefügt"}
// }

// func legacy_InitWebRTCAPI(logChan chan t.Logg) (*webrtc.API, *mediadevices.CodecSelector, error) {
// 	logChan <- t.Logg{Text: "WebRTC: InitWebRTCAPI gestartet"}
// 	// opus audio codec konfiguration
// 	opusParams, err := opus.NewParams()
// 	if err != nil {
// 		logChan <- t.Logg{Text: fmt.Sprintf("WebRTC: Fehler bei opus.NewParams: %v", err)}
// 		return nil, nil, err
// 	}

// 	codeSelector := mediadevices.NewCodecSelector(mediadevices.WithAudioEncoders(&opusParams))
// 	// verwaltet codecs
// 	mediaEngine := webrtc.MediaEngine{}
// 	// trägt codecs ein
// 	codeSelector.Populate(&mediaEngine)
// 	// einheitliche API Instanz damit alle Peers die gleichen Codecs nutzen
// 	api := webrtc.NewAPI(webrtc.WithMediaEngine(&mediaEngine))

// 	logChan <- t.Logg{Text: "WebRTC: InitWebRTCAPI abgeschlossen"}
// 	return api, codeSelector, nil
// }

// func legacy_createPeerConnection(api *webrtc.API, logChan chan t.Logg) (*webrtc.PeerConnection, error) {
// 	logChan <- t.Logg{Text: "WebRTC: createPeerConnection gestartet"}
// 	// Define ICE servers
// 	config := webrtc.Configuration{
// 		ICEServers: []webrtc.ICEServer{
// 			{
// 				URLs: []string{"stun:stun.l.google.com:19302"},
// 			},
// 		},
// 	}

// 	// Neue Peer Connection
// 	peerConnection, err := api.NewPeerConnection(config)
// 	if err != nil {
// 		logChan <- t.Logg{Text: fmt.Sprintf("WebRTC: Fehler bei NewPeerConnection: %v", err)}
// 		return nil, err
// 	}

// 	// Set the handler for ICE connection state
// 	// This will notify you when the peer has connected/disconnected
// 	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
// 		logChan <- t.Logg{Text: fmt.Sprintf("WebRTC: ICE connection state has changed %s", connectionState.String())}
// 	})

// 	logChan <- t.Logg{Text: "WebRTC: createPeerConnection abgeschlossen"}

// 	return peerConnection, nil
// }
