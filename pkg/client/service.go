package client

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	tokenGenerator "github.com/F4c3hugg3r/Go-Chat-Server/pkg/shared"
)

func NewClient() *Client {
	return &Client{
		clientId:   tokenGenerator.GenerateSecureToken(32),
		reader:     bufio.NewReader(os.Stdin),
		writer:     io.Writer(os.Stdout),
		httpClient: &http.Client{},
	}
}

// PostMessage sends a POST request to the endpoint, containing a message, read from the stdin
func (c *Client) PostMessage(url string) (int, error) {
	quit := 0
	parameteredUrl := fmt.Sprintf("%s/users/%s/message", url, c.clientId)

	input, err := c.reader.ReadString('\n')
	if err != nil {
		fmt.Printf("wrong input: %s", input)
		return quit, err
	}

	fmt.Printf("\033[1A\033[K")

	message := Message{Content: input}
	json, err := json.Marshal(message)
	if err != nil {
		fmt.Printf("wrong input: %s", json)
		return quit, err
	}

	req, err := http.NewRequest("POST", parameteredUrl, bytes.NewReader(json))
	if err != nil {
		log.Println("Fehler beim Erstellen der POST req: ", err)
		return quit, err
	}

	req.Header.Add("Authorization", c.authToken)
	req.Header.Add("Content-Type", "application/json")

	res, err := c.httpClient.Do(req)
	if err != nil {
		log.Println("Fehler beim Absenden der Nachricht: ", err)
		return quit, err
	}
	defer res.Body.Close()

	if input == "quit\n" {
		quit = 1
		fmt.Println("du hast den Channel verlassen")
	}

	return quit, nil
}

// GetMessages sends a GET request to the endpoint, displaying incoming messages
func (c *Client) GetMessages(url string) {
	parameteredUrl := fmt.Sprintf("%s/users/%s/chat", url, c.clientId)

	req, err := http.NewRequest("GET", parameteredUrl, nil)
	if err != nil {
		log.Println("Fehler beim erstellen der GET request: ", err)
	}

	req.Header.Add("Authorization", c.authToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Println("Fehler beim Abrufen ist aufgetreten: ", err)
	}

	body, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return
	}

	msg := Message{}
	dec := json.NewDecoder(strings.NewReader(string(body)))
	err = dec.Decode(&msg)
	if err != nil {
		return
	}

	messageString := msg.Name + ": " + msg.Content
	if msg.Content != "" {
		fmt.Fprint(c.writer, messageString)
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

	message := Message{Name: clientName}
	json, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("wrong input: %s", json)
	}

	resp, err := c.httpClient.Post(parameteredUrl, "application/json", bytes.NewReader(json))
	if err != nil {
		fmt.Println("Die Registrierung hat nicht funktioniert, versuch es nochmal mit anderen Daten")
		return err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Fehler beim Lesen des Bodies ist aufgetreten: ", err)
	}
	defer resp.Body.Close()
	c.authToken = string(body)

	fmt.Println("- Du wurdest registriert. -\n-> Gebe 'quit' ein, um den Chat zu verlassen\n-> Oder /help um Commands auzuführen")
	return nil
}
