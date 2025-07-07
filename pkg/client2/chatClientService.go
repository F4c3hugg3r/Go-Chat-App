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

	go chatClient.MessageReceiver(server)

	return chatClient
}

func (c *ChatClient) MessageReceiver(url string) {
	for {
		c.CheckRegistered()

		body, err := c.getMessages(url)
		if err != nil {
			continue
		}

		rsp, err := DecodeToResponse(body)
		if err != nil {
			continue
		}

		valid := c.checkResponse(&rsp)
		if valid {
			c.Output <- &rsp
		}
	}
}

func (c *ChatClient) getMessages(url string) ([]byte, error) {
	res, err := c.GetRequest(url)
	if err != nil {
		c.unregister()
		log.Printf("%v: the connection to the server couldn't be established", err)

		return nil, fmt.Errorf("%w: Fehler beim Abrufen ist aufgetreten: ", err)

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

func (c *ChatClient) checkResponse(rsp *Response) bool {
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

func (c *ChatClient) CheckRegistered() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for !c.Registered {
		c.Cond.Wait()
	}
}

func (c *ChatClient) register(body []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	rsp, err := DecodeToResponse(body)
	if err != nil {
		return fmt.Errorf("%w: Fehler beim Lesen des Bodies ist aufgetreten: ", err)
	}

	c.clientName = rsp.Name
	c.authToken = rsp.Content

	c.Registered = true
	c.Cond.Signal()

	fmt.Println("- Du wurdest registriert. -\n-> Gebe '/quit' ein, um den Chat zu verlassen\n-> Oder /help um Commands auzuführen")

	return nil
}

func (c *ChatClient) unregister() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.authToken = ""
	c.clientName = ""
	c.Registered = false

	fmt.Println("- Du bist nun vom Server getrennt.")
}

func MessageToJson(msg *Message) ([]byte, error) {
	body := []byte{}
	body, err := json.Marshal(&msg)
	if err != nil {
		return nil, fmt.Errorf("%w: error parsing json", err)
	}
	return body, nil
}

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

	err = c.register(resBody)
	if err != nil {

		return fmt.Errorf("%w: error registering client", err)
	}

	return nil
}

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
		return nil //fmt.Errorf("%s: message couldn't be send", res.Status)
	}

	return nil
}

func (c *ChatClient) PollMessages() []*Response {
	result := []*Response{}
	for {
		select {
		case msg, ok := <-c.Output:
			if !ok {
				return result
			}
			result = append(result, msg)
		default:
			return result
		}
	}
}

// DecodeToResponse decodes a responseBody to a Response struct
func DecodeToResponse(body []byte) (Response, error) {
	response := Response{}
	dec := json.NewDecoder(strings.NewReader(string(body)))

	err := dec.Decode(&response)
	if err != nil {
		return response, err
	}

	return response, nil
}

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
