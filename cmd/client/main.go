package main

import (
	"bufio"
	"crypto/rand"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

var (
	clientId   string        = generateSecureToken(32)
	reader     *bufio.Reader = bufio.NewReader(os.Stdin)
	authToken  string
	httpClient = &http.Client{}
)

func main() {
	var port = flag.Int("port", 8080, "HTTP Server Port")
	flag.Parse()
	url := fmt.Sprintf("http://localhost:%d", *port)

	if err := register(url); err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			getMessages(url)
		}
	}()

	for {
		postMessage(url)
	}
}

// postMessage sends a POST request to the endpoint, containing a message, read from the stdin
func postMessage(url string) {
	parameteredUrl := fmt.Sprintf("%s/users/%s/message", url, clientId)

	message, err := reader.ReadString('\n')
	fmt.Printf("\033[1A\033[K")
	if err != nil {
		fmt.Println("wrong input")
		return
	}

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
}

// getMessages sends a GET request to the endpoint, displaying incoming messages
func getMessages(url string) {
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

// register reads a self given name from the stdin and sends a POST request to the endpoint
func register(url string) error {
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
		fmt.Println("Die Registrierung hat nicht funktioniert, versuch es nochmal mit anderen Daten")
		return err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Fehler beim Lesen des Bodies ist aufgetreten: ", err)
	}
	defer resp.Body.Close()
	authToken = string(body)

	fmt.Println("\nDu wurdest registriert.")
	return nil
}

// generateSecureToken generates a token containing random chars
func generateSecureToken(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(b)
}
