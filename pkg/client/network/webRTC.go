package network

import (
	"fmt"
	"sync"

	t "github.com/F4c3hugg3r/Go-Chat-Server/pkg/shared"
	"github.com/ebitengine/oto/v3"
	"github.com/pion/mediadevices"
	"github.com/pion/mediadevices/pkg/codec/opus"
	"github.com/pion/webrtc/v4"
	"golang.org/x/net/context"
	opusDec "gopkg.in/hraban/opus.v2"
)

type Peer struct {
	ICEConnected bool
	peerId       string
	ownId        string
	SignalChan   chan *t.Response
	Ctx          context.Context
	Cancel       context.CancelFunc
	chatClient   *Client
	logChannel   chan t.Log
	OnChangeChan chan t.ClientsChangeSignal
	mu           *sync.RWMutex

	peerConn      *webrtc.PeerConnection
	otoContext    *oto.Context
	players       []*oto.Player
	mediaStream   mediadevices.MediaStream
	codecSelector *mediadevices.CodecSelector
	api           *webrtc.API
	iceCandidates []webrtc.ICECandidateInit
}

// initializer function
func NewPeer(opposingId string, logChannel chan t.Log, chatClient *Client, ownId string, clientsSingalChan chan t.ClientsChangeSignal) (*Peer, error) {
	var err error
	p := &Peer{
		SignalChan:    make(chan *t.Response, 100),
		peerId:        opposingId,
		ownId:         ownId,
		logChannel:    logChannel,
		chatClient:    chatClient,
		mu:            &sync.RWMutex{},
		OnChangeChan:  clientsSingalChan,
		iceCandidates: []webrtc.ICECandidateInit{},
	}

	p.Ctx, p.Cancel = context.WithCancel(context.Background())

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

	p.peerConn.OnTrack(p.OnTrackHandler)
	p.peerConn.OnICECandidate(p.OnICECandidateHandler)
	p.peerConn.OnNegotiationNeeded(p.OnNegotiationNeededHandler)
	p.peerConn.OnICEConnectionStateChange(p.OnICEConnectionStateChangeHandler)
	p.peerConn.OnSignalingStateChange(p.OnSignalingStateChangeHandler)

	go p.pollSignals()

	return p, nil
}

// maybe TODO error returnen und behandeln
// handler functions
func (p *Peer) OnTrackHandler(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
	p.logChannel <- t.Log{Text: fmt.Sprintf("WebRTC: Empfange Track: ID=%s, Kind=%s", track.ID(), track.Kind())}
	if track.Kind() != webrtc.RTPCodecTypeAudio {
		p.logChannel <- t.Log{Text: "WebRTC: Track ist kein Audio, ignoriere"}
		return
	}

	otoCtx, ready, err := oto.NewContext(&oto.NewContextOptions{
		SampleRate:   48000,
		ChannelCount: 1,
		Format:       oto.FormatSignedInt16LE,
		BufferSize:   0,
	})
	if err != nil {
		p.logChannel <- t.Log{Text: fmt.Sprintf("WebRTC: oto.NewContext error: %v", err)}
		return
	}

	p.otoContext = otoCtx
	<-ready
	p.logChannel <- t.Log{Text: fmt.Sprintf("WebRTC: oto.Context bereit %v", err)}

	decoder, err := opusDec.NewDecoder(48000, 1)
	if err != nil {
		p.logChannel <- t.Log{Text: fmt.Sprintf("WebRTC: opus.NewDecoder error: %v", err)}
		return
	}
	// maybe TODO jitter buffer für stabilere Audioqualität
	p.logChannel <- t.Log{Text: "WebRTC: Opus Decoder bereit"}

	// RTP Pakete von Opus in PCM decodieren
	reader := t.NewOpusRTPReader(track, decoder)
	player := otoCtx.NewPlayer(reader)
	p.players = append(p.players, player)
	p.logChannel <- t.Log{Text: "WebRTC: starte Audio Player"}
	go player.Play()
}

func (p *Peer) OnICEConnectionStateChangeHandler(connectionState webrtc.ICEConnectionState) {
	var err error

	p.logChannel <- t.Log{Text: fmt.Sprintf("WebRTC: ICE connection state has changed %s", connectionState.String())}
	if connectionState == webrtc.ICEConnectionStateConnected {
		p.OnChangeChan <- t.ClientsChangeSignal{CallState: t.ConnectedFlag, OppId: p.peerId}
		msg := p.chatClient.CreateMessage(p.ownId, fmt.Sprint("/", t.ConnectedFlag), "", p.peerId)
		_, err = p.chatClient.PostMessage(msg, t.SignalWebRTC)
		if err != nil {
			p.logChannel <- t.Log{Text: fmt.Sprintf("WebRTC: Fehler beim senden des Connected Flags %v", err)}
		}
	}

	if connectionState == webrtc.ICEConnectionStateFailed {
		p.ICEConnected = false
		p.chatClient.SendSignalingError(p.peerId, "")
	}
}

func (p *Peer) OnSignalingStateChangeHandler(signalingState webrtc.SignalingState) {
	switch signalingState {
	case webrtc.SignalingStateStable:
		p.SetSignalingState(t.StableSignalFlag, true)
	case webrtc.SignalingStateHaveLocalOffer:
		p.SetSignalingState(t.OfferSignalFlag, true)
	case webrtc.SignalingStateHaveRemoteOffer:
		p.SetSignalingState(t.AnswerSignalFlag, true)
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

func (p *Peer) OnNegotiationNeededHandler() {
	// Wenn Connection besteht und Signaling State Stable ist
	if p.peerConn.ICEConnectionState() == webrtc.ICEConnectionStateConnected && p.peerConn.SignalingState() == webrtc.SignalingStateStable {
		p.logChannel <- t.Log{Text: "WebRTC: Negotiation needed: sending offer"}
		err := p.OfferConnection()
		if err != nil {
			p.logChannel <- t.Log{Text: "WebRTC: Negotiation needed: failed sending offer"}
		}
	} else {
		p.logChannel <- t.Log{Text: "WebRTC: Negotiation needed: ignore"}
	}
}

// connection functions
func (p *Peer) OfferConnection() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.SetSignalingState(t.OfferSignalFlag, true)
	p.logChannel <- t.Log{Text: fmt.Sprintf("WebRTC: SendOffer started, SignalingState: %s", p.peerConn.SignalingState())}

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

	p.logChannel <- t.Log{Text: "WebRTC: LocalDescription gesetzt"}

	msg := p.chatClient.CreateMessage(p.ownId, fmt.Sprint("/", t.OfferSignalFlag), offer.SDP, p.peerId)
	_, err = p.chatClient.PostMessage(msg, t.SignalWebRTC)
	if err != nil {
		p.logChannel <- t.Log{Text: fmt.Sprintf("WebRTC: Fehler beim senden des Offers %v", err)}
		return err
	}
	p.logChannel <- t.Log{Text: "WebRTC: Offer gesendet"}
	return nil
}

func (p *Peer) CloseConnection() {
	p.chatClient.SendSignalingError(p.peerId, "")
	if p.Cancel != nil {
		p.Cancel()
	}
	if p.peerConn != nil {
		p.peerConn.Close()
	}
	if p.mediaStream != nil {
		for _, track := range p.mediaStream.GetTracks() {
			track.Close()
		}
	}
	for _, player := range p.players {
		if player != nil {
			player.Close()
		}
	}
	if p.otoContext != nil {
		p.otoContext.Suspend()
	}
}

// signaling
func (p *Peer) pollSignals() {
	var err error
	p.logChannel <- t.Log{Text: "WebRTC: pollSignals gestartet"}
	for {
		select {
		case rsp := <-p.SignalChan:
			p.logChannel <- t.Log{Text: fmt.Sprintf("WebRTC: pollSignals - empfangen: %s\n", rsp.RspName)}
			switch rsp.RspName {
			case t.ICECandidateFlag:
				p.logChannel <- t.Log{Text: "WebRTC: ReceiveICECandidate wird ausgeführt"}
				p.ReceiveICECandidate(rsp.Content)
			case t.OfferSignalFlag:
				p.logChannel <- t.Log{Text: "WebRTC: ReceiveOffer wird ausgeführt"}
				err = p.ReceiveOffer(rsp)
				if err != nil {
					p.chatClient.SendSignalingError(p.peerId, "")
				}
			case t.AnswerSignalFlag:
				p.logChannel <- t.Log{Text: "WebRTC: ReceiveAnswer wird ausgeführt"}
				err = p.ReceiveAnswer(rsp)
				if err != nil {
					p.chatClient.SendSignalingError(p.peerId, "")
				}
			}
		case <-p.Ctx.Done():
			close(p.SignalChan)
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

	msg := p.chatClient.CreateMessage(p.ownId, fmt.Sprint("/", t.AnswerSignalFlag), answer.SDP, p.peerId)
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

// helper functions
func (p *Peer) SetSignalingState(state string, sendToServer bool) error {
	p.OnChangeChan <- t.ClientsChangeSignal{CallState: state, OppId: p.peerId}

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
	// opus audio codec konfiguration
	opusParams, err := opus.NewParams()
	if err != nil {
		p.logChannel <- t.Log{Text: fmt.Sprintf("WebRTC: Fehler bei opus.NewParams: %v", err)}
		return err
	}

	p.codecSelector = mediadevices.NewCodecSelector(mediadevices.WithAudioEncoders(&opusParams))
	mediaEngine := webrtc.MediaEngine{}
	p.codecSelector.Populate(&mediaEngine)
	// einheitliche API Instanz damit alle Peers die gleichen Codecs nutzen
	p.api = webrtc.NewAPI(webrtc.WithMediaEngine(&mediaEngine))

	p.logChannel <- t.Log{Text: "WebRTC: InitWebRTCAPI abgeschlossen"}

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

	p.iceCandidates = p.iceCandidates[:0]
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
