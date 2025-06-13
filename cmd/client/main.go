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
	//normalerweise uuid
	clientId string        = fmt.Sprintf("%d-%d", time.Now().UnixNano(), rand.Int())
	reader   *bufio.Reader = bufio.NewReader(os.Stdin)
)

// TODO Documentation
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

func postMessage(url string) {
	parameteredUrl := url + "/message?clientId=" + clientId

	message, err2 := reader.ReadString('\n')
	fmt.Printf("\033[1A\033[K")
	if err2 != nil {
		fmt.Println("wrong input")
		return
	}

	//Post
	_, err := http.Post(parameteredUrl, "texp/plain", strings.NewReader(message))
	if err != nil {
		log.Println("Fehler beim Absenden der Nachricht: ", err)
		return
	}
}

func getMessages(url string) {
	parameteredUrl := url + "/chat?clientId=" + clientId

	//Get Anfrage ausführen
	res, err := http.Get(parameteredUrl)
	if err != nil {
		log.Println("Fehler beim Abrufen ist aufgetreten: ", err)
	}
	body, err2 := io.ReadAll(res.Body)
	defer res.Body.Close()
	if err2 != nil {
		log.Println("Fehler beim Lesen des Bodies ist aufgetreten: ", err2)
	}
	fmt.Println(string(body))
}

func register(url string) error {
	//Namen Scannen
	fmt.Println("Gebe deinen Namen an:")
	clientName, err2 := reader.ReadString('\n')
	clientName = strings.ReplaceAll(clientName, "\n", "")
	if err2 != nil {
		fmt.Println("wrong input")
		return err2
	}

	parameteredUrl := url + "/user?clientId=" + clientId

	//Post
	_, err := http.Post(parameteredUrl, "text/plain", strings.NewReader(clientName))
	if err != nil {
		fmt.Println("Die Registrierung hat nicht funktioniert, versuch es nochmal mit anderen Daten")
		return err
	}
	fmt.Println("\nDu wurdest registriert.")
	return nil
}
