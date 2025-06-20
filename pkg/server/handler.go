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
	deleter     ClientDeleter
}

type ClientRegisterer func(clientId, body string) (token string, e error)

type MessageBroadcaster func(msg Message)

type ClientDeleter func(clientId string) error

func NewServerHandler(chatService *ChatService) *ServerHandler {
	return &ServerHandler{
		service:     chatService,
		registerer:  chatService.registerClient,
		broadcaster: chatService.sendBroadcast,
		deleter:     chatService.logOutClient,
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

	client, err := handler.service.getClient(clientId)
	if err != nil {
		http.Error(w, "Client not found ", http.StatusNotFound)
		return
	}

	select {
	case msg, ok := <-client.clientCh:
		if !ok {
			http.Error(w, "client already deleted", http.StatusGone)
			return
		}
		message := msg.name + ": " + msg.content
		fmt.Fprint(w, message)
		return
	case <-time.After(30 * time.Second):
		fmt.Fprintf(w, "\033[1A\033[K")
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

	if client, err := handler.service.getClient(clientId); err == nil {
		if string(body) == "quit\n" {
			handler.deleter(clientId)
			handler.broadcaster(Message{fmt.Sprint("Server message - ", client.name), "logged out!\n"})
			return
		}
		handler.broadcaster(Message{client.name, string(body)})
	} else {
		http.Error(w, "client not found", http.StatusForbidden)
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

	//handler.broadcaster(Message{fmt.Sprint("\nServer message - ", string(body)), "joined the chat!\n"})
	w.Write([]byte(token))
}

// authMiddleware checks if the authToken is fitting the token given while registry and throws
// an error if not
func (handler *ServerHandler) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		clientId := r.PathValue("clientId")
		if token == "" || clientId == "" {
			http.Error(w, "missing path parameter clientId or authToken", http.StatusBadRequest)
			return
		}

		if client, err := handler.service.getClient(clientId); err != nil || token != client.authToken {
			http.Error(w, "client does not exist or token doesn't match", http.StatusForbidden)
			return
		}

		next(w, r)
	}
}
