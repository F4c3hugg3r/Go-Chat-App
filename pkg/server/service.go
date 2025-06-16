package server

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"sync"
)

// clients who communicate with the sever
type ChatService struct {
	clients map[string]*Client
	mu      sync.RWMutex
}

func NewChatService() *ChatService {
	return &ChatService{clients: make(map[string]*Client)}
}

// inactiveClientDeleter searches for inactive clients and deletes them as well as closes their message-channel
func (s *ChatService) InactiveClientDeleter() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for clientId, client := range s.clients {
		if !client.active {
			fmt.Println("due to inactivity: deleting ", client.name)
			close(client.clientCh)
			delete(s.clients, clientId)
		}
	}
}

// authMiddleware checks if the authToken is fitting the token given while registry and throws
// an error if not
func (s *ChatService) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		clientId := r.PathValue("clientId")
		if token == "" || clientId == "" {
			http.Error(w, "missing path parameter clientId or authToken", http.StatusBadRequest)
			return
		}

		s.mu.RLock()
		if client, exists := s.clients[clientId]; !exists || token != client.authToken {
			http.Error(w, "client does not exist or token doesn't match", http.StatusForbidden)
		}
		s.mu.RUnlock()

		next(w, r)
	}
}

// registerClient safely registeres a client by creating a Client with the received values
// and putting it into the global clients map
func (s *ChatService) RegisterClient(clientId, body string) (token string, e error) {
	token = generateSecureToken(64)

	s.mu.Lock()
	if _, exists := s.clients[clientId]; exists {
		return token, fmt.Errorf("client already defined")
	}
	clientCh := make(chan Message)
	s.clients[clientId] = &Client{string(body), clientId, clientCh, true, token}
	s.mu.Unlock()
	fmt.Printf("\nNew client '%s' registered.\n", body)
	return token, nil
}

// sendBroadcast distributes an incomming message abroad all client channels if
// a client can't receive, i'ts active status is set to false
func (s *ChatService) SendBroadcast(msg Message) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.clients) <= 0 {
		fmt.Printf("\nThere are no clients registered\n")
		return
	}

	for _, client := range s.clients {
		select {
		case client.clientCh <- msg:
			fmt.Println("success")
		default:
			client.active = false
		}
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
