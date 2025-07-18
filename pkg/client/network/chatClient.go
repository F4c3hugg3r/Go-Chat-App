package network

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	t "github.com/F4c3hugg3r/Go-Chat-Server/pkg/client/types"
	"github.com/F4c3hugg3r/Go-Chat-Server/pkg/shared"
)

// ChatClient handles all network tasks
type ChatClient struct {
	clientName string
	clientId   string
	authToken  string
	HttpClient *http.Client
	Registered bool
	mu         *sync.RWMutex
	cond       *sync.Cond
	Output     chan *t.Response
	Url        string
	Endpoints  map[int]string
	groupId    string
}

// NewClient generates a ChatClient and spawns a ResponseReceiver goroutine
func NewClient(server string) *ChatClient {
	chatClient := &ChatClient{
		clientId:   shared.GenerateSecureToken(32),
		Output:     make(chan *t.Response, 10000),
		HttpClient: &http.Client{},
		Registered: false,

		mu:  &sync.RWMutex{},
		Url: server,
	}

	chatClient.Endpoints = chatClient.RegisterEndpoints(chatClient.Url)
	chatClient.cond = sync.NewCond(chatClient.mu)

	go chatClient.ResponseReceiver(server)

	return chatClient
}

// RegisterEndpoints registeres endpoint urls to the corresponding enum values
func (c *ChatClient) RegisterEndpoints(url string) map[int]string {
	endpoints := make(map[int]string)
	endpoints[t.PostRegister] = fmt.Sprintf("%s/users/%s", url, c.clientId)
	endpoints[t.PostPlugin] = fmt.Sprintf("%s/users/%s/run", url, c.clientId)
	endpoints[t.Delete] = fmt.Sprintf("%s/users/%s", url, c.clientId)
	endpoints[t.Get] = fmt.Sprintf("%s/users/%s/chat", url, c.clientId)

	return endpoints
}

// Interrupt sends a Delete to the server and closes idle connections
func (c *ChatClient) Interrupt() {
	if c.Registered {
		err := c.PostDelete(c.CreateMessage("", "/quit", "", ""))
		if err != nil {
			c.Output <- &t.Response{Err: fmt.Errorf("%w: delete could not be sent", err)}
		}
	}

	c.HttpClient.CloseIdleConnections()
}

// ResponseReceiver gets responses if client is registered
// and sends then into the output channel
func (c *ChatClient) ResponseReceiver(url string) {
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
func (c *ChatClient) checkRegistered() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for !c.Registered {
		c.cond.Wait()
	}
}

// register puts values into the client flields and sends a signal
// to unblock CheckRegister
func (c *ChatClient) Register(rsp *t.Response) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.clientName = rsp.Name
	c.authToken = rsp.Content

	c.Registered = true
	c.cond.Signal()

	return nil
}

// unregister deletes client fields and sets the Registered field to false
func (c *ChatClient) Unregister() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.authToken = ""
	c.clientName = ""
	c.Registered = false
}

// GetAuthToken returns the authToken and a bool if the token is set
func (c *ChatClient) GetAuthToken() (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.authToken == "" {
		return "", false
	}

	return c.authToken, true
}

// PostMessage marshals a Message and posts it the the given endpoint
// returning the http response and an error
func (c *ChatClient) PostMessage(msg *t.Message, endpoint int) (*t.Response, error) {
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

	rsp, err := DecodeToResponse(resBody)
	if err != nil {
		return nil, fmt.Errorf("%w: error decoding body to Response", err)
	}

	return rsp, nil
}

// PostDelete sends a DELETE Request to the delete endpoint and
// unregisteres the ChatClient
func (c *ChatClient) PostDelete(msg *t.Message) error {
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

// getResponse sends a GET Request to the server, checks the http Response
// and returns the body
func (c *ChatClient) GetResponse(url string) (*t.Response, error) {
	res, err := c.GetRequest(c.Endpoints[t.Get])
	if err != nil {
		c.Unregister()

		return &t.Response{Err: fmt.Errorf("%w: the connection to the server couldn't be established", err)},
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

	rsp, err := DecodeToResponse(body)
	if err != nil {
		return nil, fmt.Errorf("%s: error decoding body to Response", res.Status)
	}

	return rsp, nil
}

// CreateMessage creates a Message with the given parameters or
// if clientName/clientId are empty fills them with the global values of the client
func (c *ChatClient) CreateMessage(clientName string, plugin string, content string, clientId string) *t.Message {
	msg := &t.Message{}

	if clientName == "" && c.Registered {
		msg.Name = c.GetName()
	} else {
		msg.Name = clientName
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

func (c *ChatClient) GetClientId() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.clientId
}

func (c *ChatClient) GetGroupId() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.groupId
}

func (c *ChatClient) SetGroupId(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.groupId = id
}

func (c *ChatClient) UnsetGroupId() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.groupId = ""
}

// GetName returns the name of the client
func (c *ChatClient) GetName() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.clientName
}

// DecodeToResponse decodes a responseBody to a Response struct
func DecodeToResponse(body []byte) (*t.Response, error) {
	response := &t.Response{}
	dec := json.NewDecoder(strings.NewReader(string(body)))

	err := dec.Decode(&response)
	if err != nil {
		return response, err
	}

	return response, nil
}

// DecodeToGroup decodes a responseBody to a Group struct
func DecodeToGroup(body []byte) (*t.Group, error) {
	group := &t.Group{}
	dec := json.NewDecoder(strings.NewReader(string(body)))

	err := dec.Decode(&group)
	if err != nil {
		return group, err
	}

	return group, nil
}
