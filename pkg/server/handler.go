package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type ServerHandler struct {
	Service *ChatService
	Plugins *PluginRegistry
}

func NewServerHandler(chatService *ChatService, pluginReg *PluginRegistry) *ServerHandler {
	return &ServerHandler{
		Service: chatService,
		Plugins: pluginReg,
	}
}

// handleGetRequest displays a response when received and times out after 10s
// if nothing is being send
// should receive a Path Parameter with clientId in it
func (handler *ServerHandler) HandleGetRequest(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Second*10)
	defer cancel()

	if r.Method != http.MethodGet {
		http.Error(w, "only GET Requests allowed", http.StatusBadRequest)
		return
	}

	clientId := r.PathValue("clientId")
	if clientId == "" {
		http.Error(w, "missing path parameter clientId", http.StatusBadRequest)
		return
	}

	client, err := handler.Service.getClient(clientId)
	if err != nil {
		http.Error(w, "client not found ", http.StatusNotFound)
		return
	}

	select {
	case rsp, ok := <-client.clientCh:
		if !ok {
			http.Error(w, "you got deleted please register again", http.StatusGone)
			return
		}

		json, err := json.Marshal(rsp)
		if err != nil {
			http.Error(w, "error formatting response to json", http.StatusInternalServerError)
			return
		}

		_, err = w.Write(json)
		if err != nil {
			http.Error(w, "couldn't write response", http.StatusInternalServerError)
		}

		return
	case <-ctx.Done():
		fmt.Fprintf(w, "\033[1A\033[K")
		return
	}
}

// handleMessages takes an incoming POST request with a message in i'ts body and distributes it to all clients
// should receive a Path Parameter with clientId in it
// should receive the message in the request body
func (handler *ServerHandler) HandleMessages(w http.ResponseWriter, r *http.Request) {
	clientId := r.PathValue("clientId")
	if clientId == "" {
		http.Error(w, "missing path parameter clientId", http.StatusBadRequest)
		return
	}

	bodyMax := http.MaxBytesReader(w, r.Body, 1<<20)
	defer r.Body.Close()

	body, err := io.ReadAll(bodyMax)

	if err != nil {
		http.Error(w, "error reading request body", http.StatusInternalServerError)
		return
	}

	message, err := DecodeToMessage(body)
	if err != nil {
		http.Error(w, "error decoding request body", http.StatusInternalServerError)
		return
	}

	res, err := handler.Plugins.FindAndExecute(&message)
	if err != nil {
		errEcho := handler.Service.echo(clientId, res)
		if errEcho != nil {
			http.Error(w, errEcho.Error(), http.StatusInternalServerError)

			return
		}

		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	if res.Name == "authToken" {
		body, err = json.Marshal(res)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		_, err = w.Write(body)
		if err != nil {
			http.Error(w, "couldn't write response", http.StatusInternalServerError)
		}

		return
	}

	handler.Service.echo(clientId, res)
}

// authMiddleware checks if the authToken is fitting the token given while registry and throws
// an error if not
func (handler *ServerHandler) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientId := r.PathValue("clientId")

		token := r.Header.Get("Authorization")
		if token == "" || clientId == "" {
			http.Error(w, "missing path parameter clientId or authToken", http.StatusBadRequest)
			return
		}

		client, err := handler.Service.getClient(clientId)
		if err != nil || token != client.authToken {
			http.Error(w, "client does not exist or token doesn't match", http.StatusForbidden)
			return
		}

		next(w, r)
	}
}
