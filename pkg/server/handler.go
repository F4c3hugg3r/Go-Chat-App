package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type ServerHandler struct {
	service *ChatService
	plugins *PluginRegistry
}

func NewServerHandler(chatService *ChatService, pluginReg *PluginRegistry) *ServerHandler {
	return &ServerHandler{
		service: chatService,
		plugins: pluginReg,
	}
}

// handleGetRequest displays a response when received and times out after 30s
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
	case rsp, ok := <-client.clientCh:
		if !ok {
			http.Error(w, "client already deleted", http.StatusGone)
			return
		}

		json, err := json.Marshal(rsp)
		if err != nil {
			http.Error(w, "Error formatting response to json", http.StatusInternalServerError)
			return
		}

		w.Write(json)
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

	message, err := DecodeToMessage(body)
	if err != nil {
		http.Error(w, "error decoding request body", http.StatusInternalServerError)
		return
	}

	res, err := handler.plugins.FindAndExecute(&message)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if res.Name == "authToken" {
		body, err = json.Marshal(res)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		w.Write(body)
		return
	}
	handler.service.echo(clientId, res)
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
