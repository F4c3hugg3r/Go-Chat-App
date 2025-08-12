package network

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	a "github.com/F4c3hugg3r/Go-Chat-Server/pkg/client/audio"
	t "github.com/F4c3hugg3r/Go-Chat-Server/pkg/shared"
)

// Client handles all network tasks
type Client struct {
	clientName     string
	clientId       string
	authToken      string
	groupId        string
	Registered     bool
	CurrentCalling string

	mu                     *sync.RWMutex
	cond                   *sync.Cond
	Output                 chan *t.Response
	LogChan                chan t.Log
	ClientChangeSignalChan chan t.ClientsChangeSignal
	CallTimeoutChan        chan bool

	Url        string
	HttpClient *http.Client
	Endpoints  map[int]string

	PortAudioMicInput *a.PortAudioMicInput
	SpeakerOutput     *a.SpeakerOutput

	Peers map[string]*Peer
}

// NewClient generates a ChatClient and spawns a ResponseReceiver goroutine
func NewClient(server string) *Client {
	var err error
	chatClient := &Client{
		clientId:               t.GenerateSecureToken(32),
		clientName:             "",
		groupId:                "",
		authToken:              "",
		Output:                 make(chan *t.Response, 10000),
		ClientChangeSignalChan: make(chan t.ClientsChangeSignal, 10000),
		CallTimeoutChan:        make(chan bool, 100),
		HttpClient:             &http.Client{},
		Registered:             false,
		CurrentCalling:         "",

		mu:      &sync.RWMutex{},
		Url:     server,
		LogChan: make(chan t.Log, 10000),

		Peers: make(map[string]*Peer),
	}

	chatClient.Endpoints = chatClient.RegisterEndpoints(chatClient.Url)
	chatClient.cond = sync.NewCond(chatClient.mu)

	chatClient.PortAudioMicInput, err = a.InitializePortAudioMic(chatClient.LogChan)
	if err != nil {
		chatClient.LogChan <- t.Log{Text: fmt.Sprintf("%v: Microphone couldn't be initialized", err)}
	}

	chatClient.SpeakerOutput, err = a.NewSpeakerOutput(chatClient.LogChan)
	if err != nil {
		chatClient.LogChan <- t.Log{Text: fmt.Sprintf("%v: Speaker Output couldn't be initialized", err)}
	}

	go chatClient.ResponseReceiver(server)

	return chatClient
}

// RegisterEndpoints registeres endpoint urls to the corresponding enum values
func (c *Client) RegisterEndpoints(url string) map[int]string {
	endpoints := make(map[int]string)
	endpoints[t.PostRegister] = fmt.Sprintf("%s/users/%s", url, c.clientId)
	endpoints[t.PostPlugin] = fmt.Sprintf("%s/users/%s/run", url, c.clientId)
	endpoints[t.Delete] = fmt.Sprintf("%s/users/%s", url, c.clientId)
	endpoints[t.Get] = fmt.Sprintf("%s/users/%s/chat", url, c.clientId)
	endpoints[t.SignalWebRTC] = fmt.Sprintf("%s/users/%s/signal", url, c.clientId)

	return endpoints
}

// Interrupt sends a Delete to the server and closes idle connections
func (c *Client) Interrupt() {
	if c.Registered {
		err := c.PostDelete(c.CreateMessage("", "/quit", "", ""))
		if err != nil {
			c.Output <- &t.Response{Err: fmt.Sprintf("%v: delete could not be sent", err)}
		}
	}

	c.DeletePeers("", true, true)
	c.PortAudioMicInput.Stream.Close()

	c.HttpClient.CloseIdleConnections()
}

// DeletePeers deletes a Peer or all Peers out of the peers map
func (c *Client) DeletePeers(oppId string, wholeMap bool, sendToOppClient bool) {
	c.LogChan <- t.Log{Text: "DeletePeersSafely gestartet"}

	if wholeMap == false {
		c.LogChan <- t.Log{Text: fmt.Sprintf("Versuche Peer mit ID %s zu löschen", oppId)}

		peer, err := c.GetPeer(oppId)
		if err != nil {
			c.LogChan <- t.Log{Text: fmt.Sprintf("%v: Peer mit ID %s existiert nicht, Abbruch", err, oppId)}
			return
		}

		if sendToOppClient {
			c.SendSignalingError(oppId, c.GetClientId(), t.FailedConnectionFlag)
		}
		c.LogChan <- t.Log{Text: fmt.Sprintf("Peer mit ID %s gefunden, Verbindung wird geschlossen", oppId)}

		peer.CloseConnection()
		c.DeletePeerSafely(oppId)
		c.LogChan <- t.Log{Text: fmt.Sprintf("Peer mit ID %s gelöscht", oppId)}

		c.SendSignalingError(oppId, c.GetClientId(), t.RollbackDoneFlag)
		c.LogChan <- t.Log{Text: fmt.Sprintf("SignalingError für Peer %s gesendet", oppId)}

		if len(c.Peers) < 1 {
			c.ClientChangeSignalChan <- t.ClientsChangeSignal{CallState: t.NoCallFlag, OppId: c.GetClientId()}
		}
		return
	}

	c.LogChan <- t.Log{Text: "Lösche alle Peers"}
	for id, peer := range c.Peers {
		if sendToOppClient {
			c.SendSignalingError(id, c.GetClientId(), t.FailedConnectionFlag)
		}
		c.LogChan <- t.Log{Text: fmt.Sprintf("Schließe Verbindung für Peer mit ID %s", id)}

		if peer != nil {
			peer.CloseConnection()
		}
		c.DeletePeerSafely(id)
		c.LogChan <- t.Log{Text: fmt.Sprintf("Peer mit ID %s gelöscht", id)}

		c.SendSignalingError(id, c.GetClientId(), t.RollbackDoneFlag)
		c.LogChan <- t.Log{Text: fmt.Sprintf("SignalingError für Peer %s gesendet", id)}
	}
	c.LogChan <- t.Log{Text: "Alle Peers wurden gelöscht"}
	c.ClientChangeSignalChan <- t.ClientsChangeSignal{CallState: t.NoCallFlag, OppId: c.GetClientId()}
}

func (c *Client) SendSignalingError(oppId string, ownId string, content string) {
	msg := c.CreateMessage(ownId, fmt.Sprintf("/"+t.FailedConnectionFlag), content, oppId)
	_, err := c.PostMessage(msg, t.SignalWebRTC)
	if err != nil {
		c.LogChan <- t.Log{Text: fmt.Sprintf("WebRTC: Fehler beim senden des ConnectionFailedFlags %v", err)}
	}
}

// ResponseReceiver gets responses if client is registered
// and sends then into the output channel
func (c *Client) ResponseReceiver(url string) {
	for {
		c.checkRegistered()

		rsp, err := c.GetResponse(url)
		if err != nil {
			continue
		}

		c.Output <- rsp
	}
}

// checkRegistered blocks until the client is being registered
func (c *Client) checkRegistered() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for !c.Registered {
		c.cond.Wait()
	}
}

// register puts values into the client flields and sends a signal
// to unblock CheckRegister
func (c *Client) Register(rsp *t.Response) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.clientName = rsp.RspName
	c.authToken = rsp.Content

	c.Registered = true
	c.cond.Signal()

	return nil
}

// unregister deletes client fields and sets the Registered field to false
func (c *Client) Unregister() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.authToken = ""
	c.clientName = ""
	c.Registered = false
}

// GetAuthToken returns the authToken and a bool if the token is set
func (c *Client) GetAuthToken() (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.authToken == "" {
		return "", false
	}

	return c.authToken, true
}

// PostMessage marshals a Message and posts it the the given endpoint
// returning the response and an error
func (c *Client) PostMessage(msg *t.Message, endpoint int) (*t.Response, error) {
	body, err := json.Marshal(&msg)
	if err != nil {
		return nil, fmt.Errorf("%w: error parsing json", err)
	}

	res, err := c.PostRequest(c.Endpoints[endpoint], body)
	if err != nil {
		return nil, fmt.Errorf("%w: message couldn't be sent", err)
	}

	defer res.Body.Close()

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: error reading response body", err)
	}

	if len(resBody) == 0 {
		return nil, nil
	}

	rsp, err := t.DecodeToResponse(resBody)
	if err != nil {
		return nil, fmt.Errorf("%w: error decoding body to Response", err)
	}

	return rsp, nil
}

// PostDelete sends a DELETE Request to the delete endpoint and
// unregisteres the ChatClient
func (c *Client) PostDelete(msg *t.Message) error {
	body, err := json.Marshal(&msg)
	if err != nil {
		return fmt.Errorf("%w: error parsing json", err)
	}

	res, err := c.DeleteRequest(c.Endpoints[t.Delete], body)
	if err != nil {
		return fmt.Errorf("%w: delete couldn't be sent", err)
	}

	defer res.Body.Close()

	c.Unregister()

	return nil
}

// getResponse sends a GET Request to the server, checks the Response
// and returns the body
func (c *Client) GetResponse(url string) (*t.Response, error) {
	res, err := c.GetRequest(c.Endpoints[t.Get])
	if err != nil {
		c.Unregister()

		return &t.Response{Err: fmt.Sprintf("%v: the connection to the server couldn't be established", err)},
			fmt.Errorf("%w: server not available", err)
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: message couldn't be received", res.Status)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("%s: message body couldn't be read", res.Status)
	}

	rsp, err := t.DecodeToResponse(body)
	if err != nil {
		return nil, fmt.Errorf("%s: error decoding body to Response", res.Status)
	}

	return rsp, nil
}

// CreateMessage creates a Message with the given parameters or
// if clientName/clientId are empty fills them with the global values of the client
func (c *Client) CreateMessage(name string, plugin string, content string, clientId string) *t.Message {
	msg := &t.Message{}

	if name == "" && c.Registered {
		msg.Name = c.GetName()
	} else {
		msg.Name = name
	}

	if clientId == "" {
		msg.ClientId = c.GetClientId()
	} else {
		msg.ClientId = clientId
	}

	msg.Content = content
	msg.Plugin = plugin
	msg.GroupId = c.GetGroupId()

	return msg
}

func (c *Client) Mute(toMute string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var inCall int

	for _, peer := range c.Peers {
		if peer.GetConnectionState() != t.ConnectedFlag {
			continue
		}
		switch toMute {
		case t.Microphone:
			peer.MuteMic()
		case t.Speaker:
			peer.MuteSpeaker()
		}
		inCall = inCall + 1
	}

	if inCall < 1 {
		return t.ErrNoPermission
	}

	return nil
}
func (c *Client) AnswerCallInitialization(message *t.Message, answer string) error {
	c.LogChan <- t.Log{Text: "AnswerCallInitialization aufgerufen"}
	if c.GetCurrentCalling() == "" {
		c.LogChan <- t.Log{Text: "Kein aktiver Anruf, Antwort nicht möglich"}
		return fmt.Errorf("%v: you are not being called", t.ErrNoPermission)
	}

	c.LogChan <- t.Log{Text: fmt.Sprintf("Sende Antwort '%s' auf Initialisierungsanruf von %s", answer, c.GetCurrentCalling())}
	msg := c.CreateMessage(message.ClientId, fmt.Sprintf("/"+t.InitializeSignalFlag), answer, c.GetCurrentCalling())

	_, err := c.PostMessage(msg, t.SignalWebRTC)
	if err != nil {
		c.LogChan <- t.Log{Text: fmt.Sprintf("Fehler beim Senden der Antwort: %v", err)}
		return err
	}

	c.LogChan <- t.Log{Text: "Antwort erfolgreich gesendet, Anrufstatus wird zurückgesetzt"}
	c.SetCurrentCalling("")

	return nil
}

func (c *Client) HandleSignal(rsp *t.Response, initialSignal bool, accepted bool) {
	err := c.HandlePeer(rsp, initialSignal, accepted)
	if err != nil {
		if initialSignal {
			c.DeletePeers(rsp.ClientId, false, true)
			return
		}
		c.DeletePeers(rsp.ClientId, false, false)
	}
}

func (c *Client) HandlePeer(rsp *t.Response, initialSignal bool, accepted bool) error {
	peer, err := c.GetPeer(rsp.ClientId)
	c.LogChan <- t.Log{Text: "Getting Peer"}

	if err != nil || peer == nil {
		c.LogChan <- t.Log{Text: "Peer existiert noch nicht, lege peer an"}

		peer, err = NewPeer(rsp.ClientId, c.LogChan, c, c.GetClientId(), c.ClientChangeSignalChan)
		if err != nil {
			c.LogChan <- t.Log{Text: fmt.Sprintf("Peer mit id: %s konnte nicht erstellt werden, server wird informiert", rsp.ClientId)}

			return err
		}

		c.SetPeer(peer)
		c.LogChan <- t.Log{Text: fmt.Sprintf("Peer mit id: %s angelegt", rsp.ClientId)}

		if initialSignal {
			c.LogChan <- t.Log{Text: "starte OfferConnection"}

			err = peer.InitializeConnection()
			if err != nil {
				return err
			}

			return nil
		}
	}

	if accepted {
		return peer.OfferConnection()
	}

	c.LogChan <- t.Log{Text: "Response wird in den Signalchannel gepusht"}
	peer.SignalChan <- rsp

	return nil
}

func (c *Client) SetPeer(peer *Peer) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.Peers[peer.peerId] = peer
}

func (c *Client) GetPeer(id string) (*Peer, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	peer, exists := c.Peers[id]
	if !exists || peer == nil {
		return nil, fmt.Errorf("%w: peer does not exist", t.ErrNotAvailable)
	}

	c.LogChan <- t.Log{Text: fmt.Sprintf("peer was found %s", peer.peerId)}

	return peer, nil
}

func (c *Client) GetClientId() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.clientId
}

func (c *Client) GetGroupId() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.groupId
}

func (c *Client) SetGroupId(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.groupId = id
}

func (c *Client) UnsetGroupId() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.groupId = ""
}

// GetName returns the name of the client
func (c *Client) GetName() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.clientName
}

func (c *Client) GetCurrentCalling() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.CurrentCalling
}

func (c *Client) SetCurrentCalling(oppId string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.CurrentCalling = oppId
}

func (c *Client) DeletePeerSafely(oppId string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.Peers, oppId)
}
