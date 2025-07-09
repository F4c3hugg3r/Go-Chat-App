package client

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/F4c3hugg3r/Go-Chat-Server/pkg/shared"
)

// NewClient generates a ChatClient and spawns a ResponseReceiver goroutine
func NewClient(server string) *ChatClient {
	chatClient := &ChatClient{
		clientId:   shared.GenerateSecureToken(32),
		Output:     make(chan *Response, 10000),
		HttpClient: &http.Client{},
		Registered: false,

		mu:  &sync.RWMutex{},
		url: server,
	}

	chatClient.Endpoints = chatClient.RegisterEndpoints(chatClient.url)
	chatClient.Cond = sync.NewCond(chatClient.mu)

	go chatClient.ResponseReceiver(server)

	return chatClient
}

func (c *ChatClient) RegisterEndpoints(url string) map[int]string {
	endpoints := make(map[int]string)
	endpoints[postRegister] = fmt.Sprintf("%s/users/%s", url, c.clientId)
	endpoints[postPlugin] = fmt.Sprintf("%s/users/%s/run", url, c.clientId)
	endpoints[delete] = fmt.Sprintf("%s/users/%s", url, c.clientId)
	endpoints[get] = fmt.Sprintf("%s/users/%s/chat", url, c.clientId)

	return endpoints
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
		c.Cond.Wait()
	}
}

// register puts values into the client flields and sends a signal
// to unblock CheckRegister
func (c *ChatClient) register(rsp *Response) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.clientName = rsp.Name
	c.authToken = rsp.Content

	c.Registered = true
	c.Cond.Signal()

	return nil
}

// unregister deletes client fields and sets the Registered field to false
func (c *ChatClient) unregister() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.authToken = ""
	c.clientName = ""
	c.Registered = false
}

func (c *ChatClient) GetAuthToken() (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.authToken, c.Registered
}

func (c *ChatClient) PostMessage(msg *Message, endpoint int) (*Response, error) {
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
func (c *ChatClient) PostDelete(msg *Message) error {
	body, err := json.Marshal(&msg)
	if err != nil {
		return fmt.Errorf("%w: error parsing json", err)
	}

	res, err := c.DeleteRequest(c.Endpoints[delete], body)
	if err != nil {
		return fmt.Errorf("%w: delete couldn't be sent", err)
	}

	defer res.Body.Close()

	c.unregister()

	return nil
}

// getResponse sends a GET Request to the server, checks the http Response
// and returns the body
func (c *ChatClient) GetResponse(url string) (*Response, error) {
	res, err := c.GetRequest(c.Endpoints[get])
	if err != nil {
		c.unregister()

		//TODO in Output Channel pushen
		log.Printf("%v: the connection to the server couldn't be established", err)

		return nil, fmt.Errorf("%w: server not available", err)
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
func (c *ChatClient) CreateMessage(clientName string, plugin string, content string, clientId string) *Message {
	msg := &Message{}

	if clientName == "" && c.Registered {
		msg.Name = c.clientName
	} else {
		msg.Name = clientName
	}

	if clientId == "" {
		msg.ClientId = c.clientId
	} else {
		msg.ClientId = clientId
	}

	msg.Content = content
	msg.Plugin = plugin

	return msg
}

// DecodeToResponse decodes a responseBody to a Response struct
func DecodeToResponse(body []byte) (*Response, error) {
	response := &Response{}
	dec := json.NewDecoder(strings.NewReader(string(body)))

	err := dec.Decode(&response)
	if err != nil {
		return response, err
	}

	return response, nil
}
