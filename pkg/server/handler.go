package server

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

type ServerHandler struct {
	service     *ChatService
	registerer  ClientRegisterer
	broadcaster MessageBroadcaster
}

type ClientRegisterer func(clientId, body string) (token string, e error)

type MessageBroadcaster func(msg Message)

func NewServerHandler(chatService *ChatService) *ServerHandler {
	return &ServerHandler{
		service:     chatService,
		registerer:  chatService.RegisterClient,
		broadcaster: chatService.SendBroadcast,
	}
}

// handleGetRequest displays a message when received and times out after 30s
// if nothing is being send
// should receive a Path Parameter with clientId in it
func (handler *ServerHandler) HandleGetRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "only GET Requests allowed", http.StatusBadRequest)
		return
	}

	clientId := r.PathValue("clientId")
	if clientId == "" {
		http.Error(w, "missing path parameter clientId", http.StatusBadRequest)
		return
	}

	handler.service.mu.RLock()
	client, exists := handler.service.clients[clientId]
	if !exists {
		handler.service.mu.RUnlock()
		http.Error(w, "Client not found ", http.StatusNotFound)
		return
	}
	clientCh := client.clientCh
	handler.service.mu.RUnlock()

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

// handleMessages takes an incoming POST request with a message in i'ts body and distributes it to all clients
// should receive a Path Parameter with clientId in it
// should receive the message in the request body
func (handler *ServerHandler) HandleMessages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "only POST request allowed", http.StatusBadRequest)
		return
	}

	clientId := r.PathValue("clientId")
	if clientId == "" {
		http.Error(w, "missing path parameter clientId", http.StatusBadRequest)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, "error reading request body", http.StatusInternalServerError)
		return
	}

	handler.service.mu.RLock()
	if client, exists := handler.service.clients[clientId]; exists {
		name := client.name
		handler.service.mu.RUnlock()
		handler.broadcaster(Message{name, string(body)})
	} else {
		handler.service.mu.RUnlock()
		http.Error(w, "client doesn't exist", http.StatusForbidden)
		return
	}
}

// handleRegistry takes an incoming POST request and lets a client register by it's name and id
// should receive a Path Parameter with clientId in it
// should receive the self given client-name in the request body
func (handler *ServerHandler) HandleRegistry(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "only POST request allowed", http.StatusBadRequest)
		return
	}

	clientId := r.PathValue("clientId")
	if clientId == "" {
		http.Error(w, "missing path parameter clientId", http.StatusBadRequest)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil || string(body) == "" {
		http.Error(w, "request body too large or empty", http.StatusBadRequest)
		return
	}

	token, err2 := handler.registerer(clientId, string(body))
	if err2 != nil {
		http.Error(w, err2.Error(), http.StatusBadRequest)
		return
	}

	w.Write([]byte(token))
}
