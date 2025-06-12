package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	scanner *bufio.Scanner = bufio.NewScanner(os.Stdin)
	//normalerweise uuid
	clientId   string = fmt.Sprintf("%d-%d", time.Now().UnixNano(), rand.Int())
	clientName string
)

func main() {
	var port = flag.Int("port", 8080, "HTTP Server Port")
	flag.Parse()
	url := fmt.Sprintf("http://localhost:%d", port)

	if err := register(url); err == nil {
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

func postMessage(url string) {
	var message string
	fmt.Println("Deine Nachricht:")
	fmt.Scan(&message)

	//Post
	_, err := http.Post(url, "texp/plain", strings.NewReader(message))
	if err != nil {
		log.Println("Fehler beim Absenden der Nachricht: ", err)
	}
}

func getMessages(url string) {
	//Get Anfrage ausführen
	res, err := http.Get(url)
	if err != nil {
		log.Println("Fehler beim Abrufen ist aufgetreten: ", err)
	}
	body, err := io.ReadAll(res.Body)
	fmt.Println(string(body))
}

func register(url string) error {
	//Namen Scannen
	fmt.Println("Gebe deinen Namen an:")
	fmt.Scan(&clientName)

	parameteredUrl := url + "?clientId=" + clientId

	//Post
	_, err := http.Post(parameteredUrl, "text/plain", strings.NewReader(clientName))
	if err != nil {
		fmt.Println("Die Registrierung hat nicht funktioniert, versuch es nochmal mit anderen Daten")
		return err
	}
	fmt.Println("Du wurdest registriert.")
	return nil
}
