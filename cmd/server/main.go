package main

import (
	"crypto/rand"
	"encoding/base64"
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
	name      string
	clientId  string
	clientCh  chan Message
	active    bool
	authToken string
}

// Message contains the name of the sender and the message (content) itsself
type Message struct {
	name    string
	content string
}

// clients who communicate with the sever
var (
	clients map[string]*Client = make(map[string]*Client)
	mu      sync.RWMutex
)

// TODO Vorschläge
// Tests
// Abmelden
// Nachrichten an alle bei neuem user
// HTTPS
func main() {
	var port = flag.Int("port", 8080, "HTTP Server Port")
	flag.Parse()
	portString := fmt.Sprintf(":%d", *port)

	http.HandleFunc("/users/{clientId}", handleRegistry)
	http.HandleFunc("/users/{clientId}/message", authMiddleware(handleMessages))
	http.HandleFunc("/users/{clientId}/chat", authMiddleware(handleGetRequest))

	go func() {
		for {
			time.Sleep(15 * time.Second)
			inactiveClientDeleter()
		}
	}()

	fmt.Println("Server running on port:", *port)
	log.Fatal(http.ListenAndServe(portString, nil))
}

// inactiveClientDeleter searches for inactive clients and deletes them as well as closes their message-channel
func inactiveClientDeleter() {
	mu.Lock()
	defer mu.Unlock()

	for clientId, client := range clients {
		if !client.active {
			fmt.Println("due to inactivity: deleting ", client.name)
			close(client.clientCh)
			delete(clients, clientId)
		}
	}
}

// authMiddleware checks if the authToken is fitting the token given while registry and throws
// an error if not
func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		clientId := r.PathValue("clientId")
		if token == "" || clientId == "" {
			http.Error(w, "missing path parameter clientId or authToken", http.StatusBadRequest)
			return
		}

		mu.RLock()
		if client, exists := clients[clientId]; !exists || token != client.authToken {
			http.Error(w, "client does not exist or token doesn't match", http.StatusForbidden)
		}
		mu.RUnlock()

		next(w, r)
	}
}

// handleRegistry takes an incoming POST request and lets a client register by it's name and id
// should receive a Path Parameter with clientId in it
// should receive the self given client-name in the request body
func handleRegistry(w http.ResponseWriter, r *http.Request) {
	clientId := r.PathValue("clientId")
	if clientId == "" {
		http.Error(w, "missing path parameter clientId", http.StatusBadRequest)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "only POST request allowed", http.StatusBadRequest)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil || string(body) == "" {
		http.Error(w, "request body too large or empty", http.StatusBadRequest)
		return
	}

	token, err2 := registerClient(clientId, string(body))
	if err2 != nil {
		http.Error(w, err2.Error(), http.StatusBadRequest)
		return
	}

	w.Write([]byte(token))
}

// registerClient safely registeres a client by creating a Client with the received values
// and putting it into the global clients map
func registerClient(clientId, body string) (token string, e error) {
	token = generateSecureToken(64)

	mu.Lock()
	if _, exists := clients[clientId]; exists {
		return token, fmt.Errorf("client already defined")
	}
	clientCh := make(chan Message)
	clients[clientId] = &Client{string(body), clientId, clientCh, true, token}
	mu.Unlock()
	fmt.Printf("\nNew client '%s' registered.\n", body)
	return token, nil
}

// handleMessages takes an incoming POST request with a message in i'ts body and distributes it to all clients
// should receive a Path Parameter with clientId in it
// should receive the message in the request body
func handleMessages(w http.ResponseWriter, r *http.Request) {
	clientId := r.PathValue("clientId")
	if clientId == "" {
		http.Error(w, "missing path parameter clientId", http.StatusBadRequest)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "only POST request allowed", http.StatusBadRequest)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, "error reading request body", http.StatusInternalServerError)
		return
	}

	mu.RLock()
	if client, exists := clients[clientId]; exists {
		name := client.name
		mu.RUnlock()
		sendBroadcast(Message{name, string(body)})
	} else {
		mu.RUnlock()
		http.Error(w, "client doesn't exist", http.StatusForbidden)
		return
	}
}

// sendBroadcast distributes an incomming message abroad all client channels if
// a client can't receive, i'ts active status is set to false
func sendBroadcast(msg Message) {
	mu.Lock()
	defer mu.Unlock()
	if len(clients) <= 0 {
		fmt.Printf("\nThere are no clients registered\n")
		return
	}

	for _, client := range clients {
		select {
		case client.clientCh <- msg:
			fmt.Println("success")
		default:
			client.active = false
		}
	}
}

// handleGetRequest displays a message when received and times out after 30s
// if nothing is being send
// should receive a Path Parameter with clientId in it
func handleGetRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "only GET Requests allowed", http.StatusBadRequest)
		return
	}

	clientId := r.PathValue("clientId")
	if clientId == "" {
		http.Error(w, "missing path parameter clientId", http.StatusBadRequest)
		return
	}

	mu.RLock()
	client, exists := clients[clientId]
	if !exists {
		mu.RUnlock()
		http.Error(w, "Client not found ", http.StatusNotFound)
		return
	}
	clientCh := client.clientCh
	mu.RUnlock()

	select {
	case msg, ok := <-clientCh:
		if !ok {
			http.Error(w, "client already deleted", http.StatusGone)
			return
		}
		message := msg.name + ": " + msg.content
		fmt.Fprint(w, message)
		return
	case <-time.After(30 * time.Second):
		return
	}
}

// generateSecureToken generates a token containing random chars
func generateSecureToken(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(b)
}
