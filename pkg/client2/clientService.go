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
	"github.com/c-bata/go-prompt"
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

	chatClient.plugins = RegisterPlugins(chatClient)
	chatClient.Cond = sync.NewCond(chatClient.mu)

	go chatClient.receiveMessages(server)

	return chatClient
}

// done
func (c *ChatClient) receiveMessages(url string) {
	for {
		c.checkRegistered()

		res, err := c.GetRequest(url)
		if err != nil {
			log.Printf("%v: Fehler beim Abrufen ist aufgetreten: ", err)
			c.unregister()

			return
		}

		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			log.Printf("%v: message couldn't be send", res.Status)
			return
		}

		body, err := io.ReadAll(res.Body)
		defer res.Body.Close()

		if err != nil {
			return
		}

		rsp, err := DecodeToResponse(body)
		if strings.TrimSpace(rsp.Content) == "" {
			return
		}

		if err != nil {
			log.Printf("%v: Fehler beim decodieren der response aufgetreten", err)
			return
		}

		valid := c.checkResponse(&rsp)
		if valid {
			c.Output <- &rsp
		}
	}
}

func (c *ChatClient) checkResponse(rsp *Response) bool {
	if rsp.Content == "" {
		return false
	}

	if rsp.Name == inactiveFlag {
		log.Println("You got kicked out due to inactivity")
		c.unregister()

		return false
	}

	return true
}

func (c *ChatClient) checkRegistered() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for !c.Registered {
		c.Cond.Wait()
	}
}

func (c *ChatClient) register() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.Registered = true
	c.Cond.Signal()
}

func (c *ChatClient) unregister() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.Registered = false
}

//PLUGINS
// body, err := json.Marshal(&msg)
// if err != nil {
// 	log.Printf("%v: error parsing json", err)
// 	return
// }

//das von unter dem switch case hier einfügen
// switch msg.Plugin {
// case "/quit":
// 	{
// 		res, err := c.DeleteRequest(c.url, body)
// 		if err != nil {
// 			log.Printf("%v: client couldn't be deleted", err)
// 			return
// 		}

// 		defer res.Body.Close()

// 		return
// 	}
// case "/broadcast":
// 	{
// 		parameteredUrl := fmt.Sprintf("%s/users/%s/run", c.url, c.clientId)
// 		res, err := c.PostRequest(parameteredUrl, body)
// 		if err != nil {
// 			log.Printf("%v: message couldn't be send", err)
// 			return
// 		}

// 		defer res.Body.Close()

// 		if res.StatusCode != http.StatusOK {
// 			log.Printf("%s: message couldn't be send", res.Status)
// 			return
// 		}
// 	}
// case "/register":
// 	{
// 		if c.Registered == true {
// 			log.Printf("you are already registered")
// 			return
// 		}

// 		//Registrierungslogik

// 		c.register()
// 	}
// }

func (c *ChatClient) sendMessage(msg *Message) {

	err := c.plugins.FindAndExecute(msg)
	if err != nil {
		log.Printf("%v: error sending message", err)
	}
}

func (c *ChatClient) Executor(input string) {
	input = strings.TrimSuffix(input, "\n")
	if !strings.HasPrefix(input, "/") {
		c.sendMessage(&Message{Name: c.clientName, Plugin: "/broadcast", Content: input, ClientId: c.clientId})
		return
	}

	if strings.HasPrefix(input, "/private") {
		opposingClientId := strings.Fields(input)[1]
		message, _ := strings.CutPrefix(input, fmt.Sprintf("/private %s ", opposingClientId))

		c.sendMessage(&Message{Name: c.clientName, Plugin: "/private", ClientId: opposingClientId, Content: message})
		return
	}

	plugin := strings.Fields(input)[0]

	content := strings.ReplaceAll(input, plugin, "")
	content, _ = strings.CutPrefix(content, " ")

	c.sendMessage(&Message{Name: c.clientName, Plugin: plugin, Content: content, ClientId: c.clientId})
}

func (c *ChatClient) Completer(d prompt.Document) []prompt.Suggest {
	s := []prompt.Suggest{
		{Text: "/help", Description: "Zeigt die Hilfe an"},
		{Text: "/quit", Description: "Beendet das Programm"},
		{Text: "/private", Description: "Sendet eine private message"},
		{Text: "/users", Description: "Listet alle User"},
		{Text: "/time", Description: "Zeigt die aktuelle Zeit an"},
	}

	//TODO bei private weiteren vorschlag für clientids

	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
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
