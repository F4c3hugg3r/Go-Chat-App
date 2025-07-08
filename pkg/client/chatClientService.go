package client2

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

const inactiveFlag = "inactive"

// NewClient generates a ChatClient and spawns a ResponseReceiver goroutine
func NewClient(server string) *ChatClient {
	chatClient := &ChatClient{
		clientId:   shared.GenerateSecureToken(32),
		Output:     make(chan *Response, 10000),
		HttpClient: &http.Client{},
		Registered: false,
		mu:         &sync.Mutex{},
		url:        server,
	}

	chatClient.Cond = sync.NewCond(chatClient.mu)

	go chatClient.ResponseReceiver(server)

	return chatClient
}

// ResponseReceiver gets responses if client is registered
// and sends then into the output channel
func (c *ChatClient) ResponseReceiver(url string) {
	for {
		c.CheckRegistered()

		body, err := c.GetJsonResponses(url)
		if err != nil {
			continue
		}

		rsp, err := DecodeToResponse(body)
		if err != nil {
			continue
		}

		valid := c.CheckResponse(rsp)
		if valid {
			c.Output <- rsp
		}
	}
}

// CheckResponse checks if the Response is empty or if the client
// was deleted due to inactvity
func (c *ChatClient) CheckResponse(rsp *Response) bool {
	if rsp.Content == "" {
		return false
	}

	if rsp.Name == inactiveFlag {
		log.Println("you got kicked out due to inactivity")
		c.unregister()

		return false
	}

	return true
}

// CheckRegistered blocks until the client is being registered
func (c *ChatClient) CheckRegistered() {
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

	fmt.Println("- Du wurdest registriert -\n-> Gebe '/quit' ein, um den Chat zu verlassen\n-> Oder '/help' um Commands auzuführen\n-> Oder ctrl+C um das Programm zu schließen")

	return nil
}

// unregister deletes client fields and sets the Registered field to false
func (c *ChatClient) unregister() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.authToken = ""
	c.clientName = ""
	c.Registered = false

	fmt.Println("- Du bist nun vom Server getrennt -")
}

// SendRegister sends a POST Request to the register endpoint and
// registeres the ChatClient
func (c *ChatClient) SendRegister(msg *Message) error {
	body, err := json.Marshal(&msg)
	if err != nil {
		return fmt.Errorf("%w: error parsing json", err)
	}

	parameteredUrl := fmt.Sprintf("%s/users/%s", c.url, c.clientId)
	res, err := c.PostRequest(parameteredUrl, body)
	if err != nil {

		return fmt.Errorf("%w: message couldn't be send", err)
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {

		return fmt.Errorf("%w: error reading response body", err)
	}

	defer res.Body.Close()

	rsp, err := DecodeToResponse(resBody)
	if err != nil {
		return fmt.Errorf("%w: error decoding body to Response", err)
	}

	err = c.register(rsp)
	if err != nil {

		return fmt.Errorf("%w: error registering client", err)
	}

	return nil
}

// SendDelete sends a DELETE Request to the delete endpoint and
// unregisteres the ChatClient
func (c *ChatClient) SendDelete(msg *Message) error {
	body, err := json.Marshal(&msg)
	if err != nil {
		return fmt.Errorf("%w: error parsing json", err)
	}

	res, err := c.DeleteRequest(c.url, body)
	if err != nil {

		return fmt.Errorf("%w: client couldn't be deleted", err)
	}

	defer res.Body.Close()

	c.unregister()

	return nil
}

// SendPlugin sends a POST request to the corresponding endpoint
// to deliver a Message
func (c *ChatClient) SendPlugin(msg *Message) error {
	body, err := json.Marshal(&msg)
	if err != nil {
		return fmt.Errorf("%w: error parsing json", err)
	}

	parameteredUrl := fmt.Sprintf("%s/users/%s/run", c.url, c.clientId)
	res, err := c.PostRequest(parameteredUrl, body)
	if err != nil {

		return fmt.Errorf("%: message couldn't be send", err)
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		//tolerated because it triggeres message poll
		return nil
	}

	return nil
}

// getResponse sends a GET Request to the server, checks the http Response
// and returns the body
func (c *ChatClient) GetJsonResponses(url string) ([]byte, error) {
	res, err := c.GetRequest(url)
	if err != nil {
		c.unregister()
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

	return body, nil
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
