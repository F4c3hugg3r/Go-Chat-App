package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

// Client is a communication participant who has a name, unique id and
// channel to receive messages
type Client struct {
	name     string
	clientId string
	clientCh chan Message
}

// Message contains the name of the sender and the message (content) itsself
type Message struct {
	name    string
	content string
}

var (
	clients map[string]Client = make(map[string]Client)
	mu      sync.Mutex
)

func main() {
	var port = flag.Int("port", 8080, "HTTP Server Port")
	flag.Parse()
	portString := fmt.Sprintf(":%d", *port)

	//Eigentlichg Query-Param aber dafür müsste ich externe bib nutzen
	http.HandleFunc("/user", handleRegistry)
	http.HandleFunc("/message", handleMessages)
	http.HandleFunc("/chat", handleGetRequest)

	fmt.Println("Server running on port:", *port)
	log.Fatal(http.ListenAndServe(portString, nil))
}

// handleRegistry lets a client register by it's name and id
// should receive a Path Parameter with clientId in it eg "clientId?fgbIUHBVIUHDCdvw"
// should receive the self given client-name in the request body
func handleRegistry(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	clientId := queryParams.Get("clientId")

	if r.Method != http.MethodPost {
		http.Error(w, "only POST request allowed", http.StatusBadRequest)
		return
	}

	if clientId == "" {
		http.Error(w, "missing query parameter", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "error reading request body", http.StatusInternalServerError)
		return
	}

	mu.Lock()
	if _, ok := clients[clientId]; ok {
		http.Error(w, "client already defined", http.StatusBadRequest)
		return
	}

	clientCh := make(chan Message)
	clients[clientId] = Client{string(body), clientId, clientCh}
	mu.Unlock()

	fmt.Printf("\nNew client '%s' registered.", string(body))
	r.Body.Close()
}

// handleMessages takes an incoming message and distributes it to all clients
// should receive a Path Parameter with clientId in it eg "clientId?fgbIUHBVIUHDCdvw"
// should receive the message in the request body
func handleMessages(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	clientId := queryParams.Get("clientId")
	if clientId == "" {
		http.Error(w, "missing query parameter", http.StatusBadRequest)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "only POST request allowed", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "error reading request body", http.StatusInternalServerError)
		return
	}
	//den Namen brauche ich eigentlich hier nicht mit angeben
	mu.Lock()
	name := clients[clientId].name
	mu.Unlock()
	sendBroadcast(Message{name, string(body)})
	r.Body.Close()
}

// sendBroadcast distributes an incomming message abroad all client Channels
func sendBroadcast(msg Message) {
	//
	mu.Lock()
	defer mu.Unlock()
	if len(clients) <= 0 {
		fmt.Printf("\n\nThere are no clients registered")
		return
	}

	for _, client := range clients {
		select {
		case client.clientCh <- msg:
			fmt.Println("success")
		case <-time.After(3 * time.Second):
			fmt.Printf("\n\nchannel of client %s full, removing client", client.name)
			//delete(clients, client.clientId)
		}
	}
}

// handleGetRequest displays a message when received and times out after 500s
// if nothing is being send
// should receive a Path Parameter with clientId in it eg "clientId?fgbIUHBVIUHDCdvw"
func handleGetRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "only GET Requests allowed", http.StatusBadRequest)
	}

	queryParams := r.URL.Query()
	clientId := queryParams.Get("clientId")
	if clientId == "" {
		http.Error(w, "missing query parameter", http.StatusBadRequest)
		return
	}
	//allgemein auf deadlocks prüfen & user validieren bzw nur ein User darf
	// auf einen Channel hören
	//safe fuction zum client holen
	select {
	case msg := <-clients[clientId].clientCh:
		message := msg.name + ": " + msg.content
		fmt.Fprint(w, message)

		return
	case <-time.After(250 * time.Second):
		return
	}
}
