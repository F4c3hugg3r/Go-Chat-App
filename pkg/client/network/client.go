package network

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	t "github.com/F4c3hugg3r/Go-Chat-Server/pkg/shared"
)

// Client handles all network tasks
type Client struct {
	clientName string
	clientId   string
	authToken  string
	groupId    string
	Registered bool
	Output     chan *t.Response

	mu       *sync.RWMutex
	cond     *sync.Cond
	LoggChan chan t.Logg

	Url        string
	HttpClient *http.Client
	Endpoints  map[int]string

	// TODO functionallity to cleer inactive Peers
	Peers map[string]*legacy_Peer
}

// NewClient generates a ChatClient and spawns a ResponseReceiver goroutine
func NewClient(server string) *Client {
	chatClient := &Client{
		clientId:   t.GenerateSecureToken(32),
		clientName: "",
		groupId:    "",
		authToken:  "",
		Output:     make(chan *t.Response, 10000),
		HttpClient: &http.Client{},
		Registered: false,

		mu:       &sync.RWMutex{},
		Url:      server,
		LoggChan: make(chan t.Logg, 10000),

		Peers: make(map[string]*legacy_Peer),
	}

	chatClient.Endpoints = chatClient.RegisterEndpoints(chatClient.Url)
	chatClient.cond = sync.NewCond(chatClient.mu)

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

	c.DeletePeersSafely()

	c.HttpClient.CloseIdleConnections()
}

func (c *Client) DeletePeersSafely() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, peer := range c.Peers {
		peer.Close(true)
		delete(c.Peers, peer.peerId)
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

func (c *Client) HandleSignal(rsp *t.Response, loggChan chan t.Logg) error {
	loggChan <- t.Logg{Text: "Getting Peer"}
	peer, err := c.GetPeer(rsp.ClientId)
	if err != nil {
		loggChan <- t.Logg{Text: "Peer existiert nicht"}

		peer = legacy_NewPeer(rsp.ClientId, c.GetClientId(), loggChan, c)
		c.SetPeer(peer)
		loggChan <- t.Logg{Text: fmt.Sprintf("Peer mit id: %s angelegt", rsp.ClientId)}

		err := peer.JoinSession()
		if err != nil {
			return err
		}

		if rsp.RspName != t.AnswerSignalFlag && rsp.RspName != t.OfferSignalFlag && rsp.RspName != t.ICECandidateFlag {
			return nil
		}
	}
	loggChan <- t.Logg{Text: "Response wird in den Signalchannel gepusht"}

	peer.SignalChan <- rsp

	return nil
}

func (c *Client) SetPeer(peer *legacy_Peer) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.Peers[peer.peerId] = peer
}

func (c *Client) GetPeer(id string) (*legacy_Peer, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	peer, exists := c.Peers[id]
	if !exists {
		return nil, fmt.Errorf("%w: peer does not exist", t.ErrNotAvailable)
	}

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
