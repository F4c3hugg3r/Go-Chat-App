package client

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	tokenGenerator "github.com/F4c3hugg3r/Go-Chat-Server/pkg/shared"
)

var (
	clientId   string        = tokenGenerator.GenerateSecureToken(32)
	reader     *bufio.Reader = bufio.NewReader(os.Stdin)
	authToken  string
	httpClient = &http.Client{}
)

// PostMessage sends a POST request to the endpoint, containing a message, read from the stdin
func PostMessage(url string) (quit error) {
	quit = nil
	parameteredUrl := fmt.Sprintf("%s/users/%s/message", url, clientId)

	message, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("wrong input")
		return
	}

	fmt.Printf("\033[1A\033[K")

	req, err := http.NewRequest("POST", parameteredUrl, strings.NewReader(message))
	if err != nil {
		log.Println("Fehler beim Erstellen der POST req: ", err)
		return
	}

	req.Header.Add("Authorization", authToken)
	req.Header.Add("Content-Type", "text/plain")

	_, err = httpClient.Do(req)
	if err != nil {
		log.Println("Fehler beim Absenden der Nachricht: ", err)
		return
	}

	if message == "quit\n" {
		quit = fmt.Errorf("\nDu hast den Channel verlassen.")
	}

	return
}

// GetMessages sends a GET request to the endpoint, displaying incoming messages
func GetMessages(url string) {
	parameteredUrl := fmt.Sprintf("%s/users/%s/chat", url, clientId)

	req, err := http.NewRequest("GET", parameteredUrl, nil)
	if err != nil {
		log.Println("Fehler beim erstellen der GET request: ", err)
	}

	req.Header.Add("Authorization", authToken)

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Println("Fehler beim Abrufen ist aufgetreten: ", err)
	}

	body, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		log.Println("Fehler beim Lesen des Bodies ist aufgetreten: ", err)
	}
	fmt.Println(string(body))
}

// Register reads a self given name from the stdin and sends a POST request to the endpoint
func Register(url string) error {
	fmt.Println("Gebe deinen Namen an:")
	clientName, err := reader.ReadString('\n')
	clientName = strings.ReplaceAll(clientName, "\n", "")
	if err != nil {
		fmt.Println("wrong input")
		return err
	}

	parameteredUrl := fmt.Sprintf("%s/users/%s", url, clientId)

	resp, err := httpClient.Post(parameteredUrl, "text/plain", strings.NewReader(clientName))
	if err != nil {
		fmt.Println("\nDie Registrierung hat nicht funktioniert, versuch es nochmal mit anderen Daten\n")
		return err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Fehler beim Lesen des Bodies ist aufgetreten: ", err)
	}
	defer resp.Body.Close()
	authToken = string(body)

	fmt.Println("\nDu wurdest registriert. Gebe 'quit' ein, um den Chat zu verlassen\n")
	return nil
}
