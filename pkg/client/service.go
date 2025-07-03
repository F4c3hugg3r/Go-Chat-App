package client

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/F4c3hugg3r/Go-Chat-Server/pkg/shared"
)

const inactive = "inactive"

func NewClient() *Client {
	return &Client{
		clientId:   shared.GenerateSecureToken(32),
		reader:     bufio.NewReader(os.Stdin),
		writer:     io.Writer(os.Stdout),
		HttpClient: &http.Client{},
	}
}

// SendMessage sends a POST request to the endpoint, containing a message, read from the stdin
func (c *Client) SendMessage(url string, cancel context.CancelFunc, input string, wg *sync.WaitGroup, ctx context.Context) error {
	if input == "" {
		inputChan := make(chan string, 1)
		errorChan := make(chan error, 1)

		wg.Add(1)

		go func() {
			defer wg.Done()

			newInput, err := c.reader.ReadString('\n')
			if err != nil {
				errorChan <- err
				return
			}
			inputChan <- newInput
		}()

		select {
		case input = <-inputChan:
		case err := <-errorChan:
			return fmt.Errorf("%w: wrong input", err)
		case <-ctx.Done():
			return nil
		}
	}

	fmt.Printf("\033[1A\033[K")

	message := c.extractInputToMessage(input)
	body, err := json.Marshal(&message)

	if err != nil {
		return fmt.Errorf("%w: error parsing json", err)
	}

	if input == "/quit\n" {
		res, err := c.DeleteRequest(url, body)
		if err != nil {
			return fmt.Errorf("%w: client couldn't be deleted", err)
		}

		defer res.Body.Close()

		cancel()

		return nil
	}

	parameteredUrl := fmt.Sprintf("%s/users/%s/run", url, c.clientId)
	res, err := c.PostRequest(parameteredUrl, body)

	if err != nil {
		return fmt.Errorf("%w: message couldn't be send", err)
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("%s: message couldn't be send", res.Status)
	}

	return err
}

// ReceiveMessages sends a GET request to the endpoint, displaying incoming messages
func (c *Client) ReceiveMessages(url string, cancel context.CancelFunc) {
	parameteredUrl := fmt.Sprintf("%s/users/%s/chat", url, c.clientId)

	res, err := c.GetRequest(parameteredUrl)
	if err != nil {
		log.Printf("%v: Fehler beim Abrufen ist aufgetreten: ", err)
		cancel()

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

	if rsp.Content == "" {
		return
	}

	if rsp.Name == inactive {
		log.Println("You got kicked out due to inactivity")
		cancel()

		return
	}

	if strings.HasPrefix(rsp.Content, "[") {
		output, err := JSONToTable(rsp.Content)
		if err != nil {
			log.Printf("%v: Fehler beim Abrufen ist aufgetreten", err)

			return
		}

		fmt.Fprint(c.writer, output)

		return
	}

	responseString := fmt.Sprintf("%s: %s\n", rsp.Name, rsp.Content)
	fmt.Fprint(c.writer, responseString)
}

// Register reads a self given name from the stdin and sends a POST request to the endpoint
func (c *Client) Register(url string) error {
	fmt.Println("Gebe deinen Namen an:")

	clientName, err := c.reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("%w: wrong input", err)
	}

	clientName = strings.ReplaceAll(clientName, "\n", "")
	message := Message{Content: clientName, ClientId: c.clientId, Plugin: "/register"}

	json, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("wrong input: %s", json)
	}

	parameteredUrl := fmt.Sprintf("%s/users/%s", url, c.clientId)

	resp, err := c.HttpClient.Post(parameteredUrl, "application/json", bytes.NewReader(json))
	if err != nil {
		return fmt.Errorf("%w: Die Registrierung hat nicht funktioniert, versuch es nochmal mit anderen Daten", err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("%w: Fehler beim Lesen des Bodies ist aufgetreten: ", err)
	}
	defer resp.Body.Close()

	rsp, err := DecodeToResponse(body)
	if err != nil {
		return fmt.Errorf("%w: Fehler beim Lesen des Bodies ist aufgetreten: ", err)
	}

	c.authToken = rsp.Content
	c.clientName = clientName

	fmt.Println("- Du wurdest registriert. -\n-> Gebe '/quit' ein, um den Chat zu verlassen\n-> Oder /help um Commands auzuführen")

	return nil
}

// extractInputToJson creates a Message type message out of the given
// input string. If the string starts with "/text", "/text" will be the plugin
func (c *Client) extractInputToMessage(input string) *Message {
	input = strings.TrimSuffix(input, "\n")
	if !strings.HasPrefix(input, "/") {
		return &Message{Name: c.clientName, Plugin: "/broadcast", Content: input, ClientId: c.clientId}
	}

	if strings.HasPrefix(input, "/private") {
		opposingClientId := strings.Fields(input)[1]
		message, _ := strings.CutPrefix(input, fmt.Sprintf("/private %s ", opposingClientId))

		return &Message{Name: c.clientName, Plugin: "/private", ClientId: opposingClientId, Content: message}
	}

	plugin := strings.Fields(input)[0]

	content := strings.ReplaceAll(input, plugin, "")
	content, _ = strings.CutPrefix(content, " ")

	return &Message{Name: c.clientName, Plugin: plugin, Content: content, ClientId: c.clientId}
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
