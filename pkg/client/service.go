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

	"github.com/F4c3hugg3r/Go-Chat-Server/pkg/shared"
)

func NewClient() *Client {
	return &Client{
		clientId:   shared.GenerateSecureToken(32),
		reader:     bufio.NewReader(os.Stdin),
		writer:     io.Writer(os.Stdout),
		HttpClient: &http.Client{},
	}
}

// PostMessage sends a POST request to the endpoint, containing a message, read from the stdin
func (c *Client) PostMessage(url string, cancel context.CancelFunc, input string) error {
	parameteredUrl := fmt.Sprintf("%s/users/%s/run", url, c.clientId)

	var err error
	if input == "" {
		input, err = c.reader.ReadString('\n')
		if err != nil {
			fmt.Printf("wrong input: %s", input)
			return err
		}
	}

	input = strings.TrimSuffix(input, "\n")

	fmt.Printf("\033[1A\033[K")

	message := extractInputToMessageFields(input, c.clientId)
	message.Name = c.clientName
	json, err := json.Marshal(message)

	if err != nil {
		fmt.Printf("wrong input: %s", json)
		return err
	}

	if input == "/quit" {
		err = c.DeleteClient(url, json)
		if err != nil {
			return fmt.Errorf("%w: client could't be deleted", err)
		}

		cancel()

		return nil
	}

	req, err := http.NewRequest("POST", parameteredUrl, bytes.NewReader(json))
	if err != nil {
		log.Println("Fehler beim Erstellen der POST req: ", err)
		return err
	}

	req.Header.Add("Authorization", c.authToken)
	req.Header.Add("Content-Type", "application/json")

	res, err := c.HttpClient.Do(req)
	if err != nil {
		log.Println("Fehler beim Absenden der Nachricht: ", err)
		return err
	}
	defer res.Body.Close()

	return nil
}

// DeleteClient sends a DELETE Request to delete the client out of the server
func (c *Client) DeleteClient(url string, json []byte) error {
	parameteredUrl := fmt.Sprintf("%s/users/%s", url, c.clientId)
	req, err := http.NewRequest("DELETE", parameteredUrl, bytes.NewReader(json))

	if err != nil {
		log.Println("Fehler beim Erstellen der DELETE req: ", err)
		return err
	}

	req.Header.Add("Authorization", c.authToken)
	req.Header.Add("Content-Type", "application/json")

	res, err := c.HttpClient.Do(req)
	if err != nil {
		log.Println("Fehler beim Absenden des Deletes: ", err)
		return err
	}

	defer res.Body.Close()

	return nil
}

// GetMessages sends a GET request to the endpoint, displaying incoming messages
func (c *Client) GetMessages(url string, cancel context.CancelFunc) {
	parameteredUrl := fmt.Sprintf("%s/users/%s/chat", url, c.clientId)

	req, err := http.NewRequest("GET", parameteredUrl, nil)
	if err != nil {
		log.Println("Fehler beim erstellen der GET request: ", err)
		return
	}

	req.Header.Add("Authorization", c.authToken)

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		log.Println("Fehler beim Abrufen ist aufgetreten: ", err)
		cancel()

		return
	}

	body, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()

	if err != nil {
		return
	}

	rsp, err := DecodeToResponse(body)
	if err != nil {
		return
	}

	if rsp.Name == "inactive" {
		log.Println("You got kicked out due to inactivity")
		cancel()

		return
	}

	if strings.HasPrefix(rsp.Content, "[") {
		output, err := JSONToTable(rsp.Content)
		if err != nil {
			log.Println("Fehler beim Abrufen ist aufgetreten: ", err)

			return
		}

		fmt.Fprint(c.writer, output)

		return
	}

	responseString := rsp.Name + ": " + rsp.Content + "\n"
	if rsp.Content != "" {
		fmt.Fprint(c.writer, responseString)
	}
}

// Register reads a self given name from the stdin and sends a POST request to the endpoint
func (c *Client) Register(url string) error {
	fmt.Println("Gebe deinen Namen an:")

	clientName, err := c.reader.ReadString('\n')
	if err != nil {
		fmt.Println("wrong input")
		return err
	}

	clientName = strings.ReplaceAll(clientName, "\n", "")
	parameteredUrl := fmt.Sprintf("%s/users/%s", url, c.clientId)

	message := Message{Content: clientName, ClientId: c.clientId, Plugin: "/register"}
	json, err := json.Marshal(message)

	if err != nil {
		return fmt.Errorf("wrong input: %s", json)
	}

	resp, err := c.HttpClient.Post(parameteredUrl, "application/json", bytes.NewReader(json))
	if err != nil {
		fmt.Println("Die Registrierung hat nicht funktioniert, versuch es nochmal mit anderen Daten")
		return err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Fehler beim Lesen des Bodies ist aufgetreten: ", err)
		return err
	}
	defer resp.Body.Close()

	rsp, err := DecodeToResponse(body)
	if err != nil {
		log.Println("Fehler beim decoden des bodies aufgetreten: ", err)
		return err
	}

	c.authToken = rsp.Content
	c.clientName = clientName

	fmt.Println("- Du wurdest registriert. -\n-> Gebe 'quit' ein, um den Chat zu verlassen\n-> Oder /help um Commands auzuführen")

	return nil
}

// extractInputToMessageFields creates a Message type message out of the given
// input string. If the string starts with "/text", "/text" will be the plugin
func extractInputToMessageFields(input string, clientId string) Message {
	if !strings.HasPrefix(input, "/") {
		return Message{Plugin: "/broadcast", Content: input, ClientId: clientId}
	}

	if strings.HasPrefix(input, "/private") {
		opposingClientId := strings.Fields(input)[1]
		message, _ := strings.CutPrefix(input, fmt.Sprintf("/private %s ", opposingClientId))

		return Message{Plugin: "/private", ClientId: opposingClientId, Content: message}
	}

	plugin := strings.Fields(input)[0]

	content := strings.ReplaceAll(input, plugin, "")
	content, _ = strings.CutPrefix(content, " ")

	return Message{Plugin: plugin, Content: content, ClientId: clientId}
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
