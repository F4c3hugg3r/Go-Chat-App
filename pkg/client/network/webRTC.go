package network

import (
	"fmt"
	"sync"

	a "github.com/F4c3hugg3r/Go-Chat-Server/pkg/client/audio"
	t "github.com/F4c3hugg3r/Go-Chat-Server/pkg/shared"
	"github.com/ebitengine/oto/v3"
	_ "github.com/pion/mediadevices/pkg/driver/microphone"
	"github.com/pion/webrtc/v4"
	"golang.org/x/net/context"
	opusDec "gopkg.in/hraban/opus.v2"
)

//
// IMPORTANT NOTE: for WebRTC Signals, Message.Name represents the ownId and Message.ClientId represents the oppId
//

type Peer struct {
	peerId                  string
	ownId                   string
	micMuted                bool
	speakerMuted            bool
	SignalChan              chan *t.Response
	ClientsChangeSignalChan chan t.ClientsChangeSignal
	logChannel              chan t.Log
	Ctx                     context.Context
	Cancel                  context.CancelFunc
	chatClient              *Client
	peerConn                *webrtc.PeerConnection
	api                     *webrtc.API
	iceCandidates           []webrtc.ICECandidateInit
	audioTransceiver        *webrtc.RTPTransceiver
	players                 []*oto.Player
	mu                      *sync.RWMutex
	Decoder                 *opusDec.Decoder
}

// initializer functions
func NewPeer(opposingId string, logChannel chan t.Log, chatClient *Client, ownId string, clientsSingalChan chan t.ClientsChangeSignal) (*Peer, error) {
	var err error
	p := &Peer{
		SignalChan:              make(chan *t.Response, 100),
		peerId:                  opposingId,
		ownId:                   ownId,
		logChannel:              logChannel,
		chatClient:              chatClient,
		mu:                      &sync.RWMutex{},
		ClientsChangeSignalChan: clientsSingalChan,
		iceCandidates:           []webrtc.ICECandidateInit{},
		micMuted:                false,
	}

	p.Ctx, p.Cancel = context.WithCancel(context.Background())
	p.logChannel <- t.Log{Text: fmt.Sprintf("WebRTC: Peer-Objekt initialisiert für OpposingId=%s, OwnId=%s", opposingId, ownId)}

	p.logChannel <- t.Log{Text: "WebRTC: Peer erstellt, starte InitWebRTC"}
	err = p.InitWebRTCAPI()
	if err != nil {
		p.logChannel <- t.Log{Text: fmt.Sprintf("WebRTC: Fehler bei InitWebRTCAPI: %v", err)}
		return nil, err
	}

	err = p.CreatePeerConnection()
	if err != nil {
		p.logChannel <- t.Log{Text: fmt.Sprintf("WebRTC: Fehler bei createPeerConnection: %v", err)}
		return nil, err
	}
	p.logChannel <- t.Log{Text: "WebRTC: PeerConnection erfolgreich erstellt"}

	p.peerConn.OnTrack(p.OnTrackHandler)
	p.peerConn.OnICECandidate(p.OnICECandidateHandler)
	// erstmal nicht gewollt
	// p.peerConn.OnNegotiationNeeded(p.OnNegotiationNeededHandler)
	p.peerConn.OnICEConnectionStateChange(p.OnICEConnectionStateChangeHandler)
	p.peerConn.OnSignalingStateChange(p.OnSignalingStateChangeHandler)
	p.peerConn.OnICEGatheringStateChange(p.OnICEGatheringStateChangeHandler)
	p.logChannel <- t.Log{Text: "WebRTC: Event-Handler für PeerConnection registriert"}

	err = p.AddTracksToPeerConnection(p.logChannel)
	if err != nil {
		p.logChannel <- t.Log{Text: fmt.Sprintf("WebRTC: Fehler beim Hinzufügen von Tracks: %v", err)}
	}

	p.logChannel <- t.Log{Text: "WebRTC: Tracks erfolgreich hinzugefügt"}

	go p.pollSignals()
	p.logChannel <- t.Log{Text: "WebRTC: pollSignals Goroutine gestartet"}

	return p, nil
}

func (p *Peer) CreatePeerConnection() error {
	var err error
	p.logChannel <- t.Log{Text: "WebRTC: createPeerConnection gestartet"}
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}

	p.peerConn, err = p.api.NewPeerConnection(config)
	if err != nil {
		p.logChannel <- t.Log{Text: fmt.Sprintf("WebRTC: Fehler bei NewPeerConnection: %v", err)}
		return err
	}

	p.logChannel <- t.Log{Text: "WebRTC: createPeerConnection abgeschlossen"}

	return nil
}

func (p *Peer) InitWebRTCAPI() error {
	p.logChannel <- t.Log{Text: "WebRTC: InitWebRTCAPI gestartet"}

	mediaEngine := webrtc.MediaEngine{}
	p.chatClient.PortAudioMicInput.CodecSelector.Populate(&mediaEngine)
	p.api = webrtc.NewAPI(webrtc.WithMediaEngine(&mediaEngine))

	p.logChannel <- t.Log{Text: "WebRTC: InitWebRTCAPI abgeschlossen"}

	return nil
}

func (p *Peer) AddTracksToPeerConnection(logChannel chan t.Log) error {
	var err error
	p.audioTransceiver, err = p.peerConn.AddTransceiverFromTrack(p.chatClient.PortAudioMicInput.Track,
		webrtc.RTPTransceiverInit{
			Direction: webrtc.RTPTransceiverDirectionSendrecv,
		},
	)
	if err != nil {
		logChannel <- t.Log{Text: fmt.Sprintf("WebRTC: Fehler bei AddTransceiverFromTrack: %v", err)}
		return err
	}

	err = p.chatClient.PortAudioMicInput.Stream.Start()
	if err != nil {
		p.logChannel <- t.Log{Text: fmt.Sprintf("PortAudio: Fehler beim Starten des Streams: %v", err)}
	}

	return nil
}

// handler functions
func (p *Peer) OnICEGatheringStateChangeHandler(state webrtc.ICEGatheringState) {
	p.logChannel <- t.Log{Text: fmt.Sprintf("WebRTC: ICEGatheringState has changed to %s", state.String())}
}

func (p *Peer) OnTrackHandler(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
	var err error
	p.logChannel <- t.Log{Text: fmt.Sprintf("WebRTC: Empfange Track: ID=%s, Kind=%s", track.ID(), track.Kind())}
	if track.Kind() != webrtc.RTPCodecTypeAudio {
		p.logChannel <- t.Log{Text: "WebRTC: Track ist kein Audio, ignoriere"}
		return
	}

	p.Decoder, err = opusDec.NewDecoder(48000, 1)
	if err != nil {
		p.logChannel <- t.Log{Text: fmt.Sprintf("WebRTC: opus.NewDecoder error: %v", err)}
		return
	}
	// maybe TODO jitter buffer für stabilere Audioqualität
	p.logChannel <- t.Log{Text: "WebRTC: Opus Decoder bereit"}

	// RTP Pakete von Opus in PCM decodieren
	reader := a.NewOpusRTPReader(track, p.Decoder)
	player := p.chatClient.SpeakerOutput.OtoContext.NewPlayer(reader)
	p.players = append(p.players, player)
	p.logChannel <- t.Log{Text: "WebRTC: starte Audio Player"}
	go player.Play()
}

func (p *Peer) OnICEConnectionStateChangeHandler(connectionState webrtc.ICEConnectionState) {
	var err error

	p.logChannel <- t.Log{Text: fmt.Sprintf("WebRTC: ICE connection state has changed %s", connectionState.String())}
	if connectionState == webrtc.ICEConnectionStateConnected {
		p.ClientsChangeSignalChan <- t.ClientsChangeSignal{CallState: t.ConnectedFlag, OppId: p.peerId}
		p.ClientsChangeSignalChan <- t.ClientsChangeSignal{CallState: t.ConnectedFlag, OppId: p.ownId}
		msg := p.chatClient.CreateMessage(p.ownId, fmt.Sprint("/", t.ConnectedFlag), "", p.peerId)
		_, err = p.chatClient.PostMessage(msg, t.SignalWebRTC)
		if err != nil {
			p.logChannel <- t.Log{Text: fmt.Sprintf("WebRTC: Fehler beim senden des Connected Flags %v", err)}
		}
	}

	if connectionState == webrtc.ICEConnectionStateFailed || connectionState == webrtc.ICEConnectionStateClosed {
		p.chatClient.SendSignalingError(p.peerId, p.ownId, "")
	}
}

func (p *Peer) OnSignalingStateChangeHandler(signalingState webrtc.SignalingState) {
	switch signalingState {
	case webrtc.SignalingStateStable:
		p.SetSignalingState(t.StableSignalFlag, true)
		p.ClientsChangeSignalChan <- t.ClientsChangeSignal{CallState: t.StableSignalFlag, OppId: p.ownId}
	case webrtc.SignalingStateHaveLocalOffer:
		p.SetSignalingState(t.OfferSignalFlag, true)
		p.ClientsChangeSignalChan <- t.ClientsChangeSignal{CallState: t.OfferSignalFlag, OppId: p.ownId}
	case webrtc.SignalingStateHaveRemoteOffer:
		p.SetSignalingState(t.AnswerSignalFlag, true)
		p.ClientsChangeSignalChan <- t.ClientsChangeSignal{CallState: t.AnswerSignalFlag, OppId: p.ownId}
	}
}

func (p *Peer) OnICECandidateHandler(candidate *webrtc.ICECandidate) {
	if candidate == nil {
		p.logChannel <- t.Log{Text: "WebRTC: ICECandidate ist nil"}
		return
	}

	candidateMsg := p.chatClient.CreateMessage(p.ownId, fmt.Sprint("/", t.ICECandidateFlag), candidate.ToJSON().Candidate, p.peerId)
	_, err := p.chatClient.PostMessage(candidateMsg, t.SignalWebRTC)
	if err != nil {
		p.logChannel <- t.Log{Text: fmt.Sprintf("WebRTC: Fehler beim Senden des ICE Candidates: %v", err)}
	}
}

// func (p *Peer) OnNegotiationNeededHandler() {
// 	// Wenn Connection besteht und Signaling State Stable ist
// 	if p.peerConn.ICEConnectionState() == webrtc.ICEConnectionStateConnected && p.peerConn.SignalingState() == webrtc.SignalingStateStable {
// 		p.logChannel <- t.Log{Text: "WebRTC: Negotiation needed: sending offer"}
// 		err := p.OfferConnection()
// 		if err != nil {
// 			p.logChannel <- t.Log{Text: "WebRTC: Negotiation needed: failed sending offer"}
// 		}
// 	} else {
// 		p.logChannel <- t.Log{Text: "WebRTC: Negotiation needed: ignore"}
// 	}
// }

// connection functions
func (p *Peer) InitializeConnection() error {
	p.SetSignalingState(t.OfferSignalFlag, false)

	msg := p.chatClient.CreateMessage(p.ownId, fmt.Sprint("/", t.InitializeSignalFlag), "", p.peerId)
	_, err := p.chatClient.PostMessage(msg, t.SignalWebRTC)
	if err != nil {
		p.logChannel <- t.Log{Text: fmt.Sprintf("WebRTC: Fehler beim senden des Offers %v", err)}
		return err
	}
	p.logChannel <- t.Log{Text: "WebRTC: Initialization Offer gesendet"}
	return nil
}

func (p *Peer) OfferConnection() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.logChannel <- t.Log{Text: "WebRTC: Erstelle SDP Offer"}

	offer, err := p.peerConn.CreateOffer(nil)
	if err != nil {
		p.logChannel <- t.Log{Text: fmt.Sprintf("WebRTC: Fehler bei CreateOffer: %v", err)}
		return err
	}

	p.logChannel <- t.Log{Text: "WebRTC: SDP Offer erstellt"}

	err = p.peerConn.SetLocalDescription(offer)
	if err != nil {
		p.logChannel <- t.Log{Text: fmt.Sprintf("WebRTC: Fehler bei SetLocalDescription: %v", err)}
		return err
	}

	p.logChannel <- t.Log{Text: "WebRTC: LocalDescription gesetzt, waiting for ICE gathering to start"}

	msg := p.chatClient.CreateMessage(p.ownId, fmt.Sprint("/", t.OfferSignalFlag), p.peerConn.LocalDescription().SDP, p.peerId)
	_, err = p.chatClient.PostMessage(msg, t.SignalWebRTC)
	if err != nil {
		p.logChannel <- t.Log{Text: fmt.Sprintf("WebRTC: Fehler beim senden des Offers %v", err)}
		return err
	}
	p.logChannel <- t.Log{Text: "WebRTC: Offer gesendet"}
	return nil
}

func (p *Peer) MuteMic() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	switch p.micMuted {
	case false:
		p.chatClient.PortAudioMicInput.Stream.Stop()
		p.micMuted = true
		p.logChannel <- t.Log{Text: "Microphone muted"}

	case true:
		p.chatClient.PortAudioMicInput.Stream.Start()
		p.micMuted = false
		p.logChannel <- t.Log{Text: "Microphone unmuted"}
	}

	return nil
}

func (p *Peer) MuteSpeaker() {
	switch p.micMuted {
	case false:
		for _, player := range p.players {
			if player != nil {
				player.Pause()
				p.logChannel <- t.Log{Text: "WebRTC: Speaker muted"}
			}
		}
		p.micMuted = true

	case true:
		for _, player := range p.players {
			if player != nil {
				player.Play()
				p.logChannel <- t.Log{Text: "WebRTC: Speaker unmuted"}
			}
		}
		p.micMuted = false
	}
}

func (p *Peer) CloseConnection() {
	p.logChannel <- t.Log{Text: "WebRTC: CloseConnection gestartet"}
	if p.Cancel != nil {
		p.logChannel <- t.Log{Text: "WebRTC: Cancel context"}
		p.Cancel()
	}

	p.logChannel <- t.Log{Text: "WebRTC: Closing SignalChan"}
	close(p.SignalChan)

	for i, player := range p.players {
		if player != nil {
			p.logChannel <- t.Log{Text: fmt.Sprintf("WebRTC: Closing Player #%d", i)}
			player.Close()
			p.logChannel <- t.Log{Text: fmt.Sprintf("WebRTC: Player #%d geschlossen", i)}
		}
	}

	if p.peerConn != nil {
		p.logChannel <- t.Log{Text: "WebRTC: Closing PeerConnection"}
		err := p.peerConn.Close()
		if err != nil {
			p.logChannel <- t.Log{Text: fmt.Sprintf("WebRTC: Fehler beim Schließen der PeerConnection: %v", err)}
		} else {
			p.logChannel <- t.Log{Text: "WebRTC: PeerConnection erfolgreich geschlossen"}
		}
	}

	p.logChannel <- t.Log{Text: "WebRTC: CloseConnection abgeschlossen"}
}

func (p *Peer) GetConnectionState() string {
	switch p.peerConn.ConnectionState() {
	case webrtc.PeerConnectionStateConnected:
		return t.ConnectedFlag
	default:
		return t.FailedConnectionFlag
	}
}

// signaling
func (p *Peer) pollSignals() {
	var err error
	p.logChannel <- t.Log{Text: "WebRTC: pollSignals gestartet"}
	for {
		select {
		case rsp := <-p.SignalChan:
			if rsp == nil {
				p.logChannel <- t.Log{Text: "WebRTC: pollSignals - Warnung: nil-Response empfangen!"}
				continue
			}
			p.logChannel <- t.Log{Text: fmt.Sprintf("WebRTC: pollSignals - empfangen: %s\n", rsp.RspName)}
			switch rsp.RspName {
			case t.ICECandidateFlag:
				p.logChannel <- t.Log{Text: "WebRTC: ReceiveICECandidate wird ausgeführt"}
				p.ReceiveICECandidate(rsp.Content)
			case t.OfferSignalFlag:
				p.logChannel <- t.Log{Text: "WebRTC: ReceiveOffer wird ausgeführt"}
				err = p.ReceiveOffer(rsp)
				if err != nil {
					p.logChannel <- t.Log{Text: fmt.Sprintf("WebRTC: Fehler in ReceiveOffer: %v", err)}
					p.chatClient.SendSignalingError(p.peerId, p.ownId, "")
				}
			case t.AnswerSignalFlag:
				p.logChannel <- t.Log{Text: "WebRTC: ReceiveAnswer wird ausgeführt"}
				err = p.ReceiveAnswer(rsp)
				if err != nil {
					p.logChannel <- t.Log{Text: fmt.Sprintf("WebRTC: Fehler in ReceiveAnswer: %v", err)}
					p.chatClient.SendSignalingError(p.peerId, p.ownId, "")
				}
			}
		case <-p.Ctx.Done():
			// p.logChannel <- t.Log{Text: "WebRTC: pollSignals - Kontext beendet, Channel wird geschlossen"}
			// close(p.SignalChan)
			p.logChannel <- t.Log{Text: "WebRTC: pollSignals - Kontext beendet"}
			return
		}
	}
}

// receiver functions
func (p *Peer) ReceiveOffer(rsp *t.Response) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	err := p.peerConn.SetRemoteDescription(webrtc.SessionDescription{
		Type: webrtc.SDPTypeOffer,
		SDP:  rsp.Content,
	})
	if err != nil {
		p.logChannel <- t.Log{Text: fmt.Sprintf("WebRTC: Fehler beim SetRemoteDescription (%s): %v", webrtc.SDPTypeOffer.String(), err)}
		return err
	}

	p.processPendingICECandidates()

	p.logChannel <- t.Log{Text: fmt.Sprintf("WebRTC: RemoteDescription (%s) gesetzt", webrtc.SDPTypeOffer.String())}

	answer, err := p.peerConn.CreateAnswer(nil)
	if err != nil {
		p.logChannel <- t.Log{Text: fmt.Sprintf("WebRTC: Fehler beim CreateAnswer: %v", err)}
		return err
	}

	err = p.peerConn.SetLocalDescription(answer)
	if err != nil {
		p.logChannel <- t.Log{Text: fmt.Sprintf("WebRTC: Fehler beim SetLocalDescription (Answer): %v", err)}
		return err
	}

	p.logChannel <- t.Log{Text: "WebRTC: LocalDescription (Answer) gesetzt"}

	msg := p.chatClient.CreateMessage(p.ownId, fmt.Sprint("/", t.AnswerSignalFlag), p.peerConn.LocalDescription().SDP, p.peerId)
	_, err = p.chatClient.PostMessage(msg, t.SignalWebRTC)
	if err != nil {
		p.logChannel <- t.Log{Text: fmt.Sprintf("WebRTC: Fehler beim Senden der Signal-Answer: %v", err)}
		return err
	}

	p.logChannel <- t.Log{Text: "WebRTC: Signal Answer gesendet"}
	return nil
}

func (p *Peer) ReceiveAnswer(rsp *t.Response) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	err := p.peerConn.SetRemoteDescription(webrtc.SessionDescription{
		Type: webrtc.SDPTypeAnswer,
		SDP:  rsp.Content,
	})
	if err != nil {
		p.logChannel <- t.Log{Text: fmt.Sprintf("WebRTC: Fehler beim SetRemoteDescription (%s): %v", webrtc.SDPTypeAnswer.String(), err)}
		return err
	}

	p.processPendingICECandidates()

	p.logChannel <- t.Log{Text: fmt.Sprintf("WebRTC: RemoteDescription (%s) gesetzt", webrtc.SDPTypeAnswer.String())}
	return nil
}

func (p *Peer) ReceiveICECandidate(ICECandidate string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.logChannel <- t.Log{Text: "WebRTC: HandleIncomingCandidate gestartet"}
	candidateInit := webrtc.ICECandidateInit{Candidate: ICECandidate}
	p.iceCandidates = append(p.iceCandidates, candidateInit)

	if p.peerConn.RemoteDescription() == nil {
		p.logChannel <- t.Log{Text: "WebRTC: ICECandidate gespeichert - keine RemoteDescription"}
		return
	}

	p.processPendingICECandidates()
}

func (p *Peer) SetSignalingState(state string, sendToServer bool) error {
	p.logChannel <- t.Log{Text: "WebRTC: SetSignalingState gestartet, changeSignal gesendet"}
	p.ClientsChangeSignalChan <- t.ClientsChangeSignal{CallState: state, OppId: p.peerId}

	if sendToServer {
		msg := p.chatClient.CreateMessage(p.ownId, fmt.Sprint("/", state), "", p.peerId)
		_, err := p.chatClient.PostMessage(msg, t.SignalWebRTC)
		if err != nil {
			p.logChannel <- t.Log{Text: fmt.Sprintf("WebRTC: Fehler beim senden der %s: %v", state, err)}
			return err
		}
	}

	return nil
}

func (p *Peer) processPendingICECandidates() {
	for _, candidate := range p.iceCandidates {
		err := p.peerConn.AddICECandidate(candidate)
		if err != nil {
			p.logChannel <- t.Log{Text: fmt.Sprintf("WebRTC: Fehler beim AddICECandidate: %v", err)}
		} else {
			p.logChannel <- t.Log{Text: "WebRTC: ICECandidate hinzugefügt"}
		}
	}

	clear(p.iceCandidates)
}

// Funktion kann wahrscheinlich gelöscht werden, da kein rollback bei fehler sondern kompletter Abbruch

// func (p *Peer) Rollback() error {
// 	p.mu.Lock()
// 	defer p.mu.Unlock()

// 	p.SetSignalingState(t.RollbackSignalFlag)

// 	p.CloseConnection()

// 	var err error
// 	p, err = NewPeer(p.peerId, p.logChannel, p.chatClient)
// 	if err != nil {
// 		p.logChannel <- t.Logg{Text: fmt.Sprintf("WebRTC: Fehler beim Erstellen eines neuen Peers %v", err)}
// 		return err
// 	}

// 	// mglw redundant
// 	p.SetSignalingState(t.StableSignalFlag)

// 	p.logChannel <- t.Logg{Text: "WebRTC: Signal Answer gesendet"}
// 	return nil
// }
