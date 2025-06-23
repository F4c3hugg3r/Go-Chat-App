package server

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	tokenGenerator "github.com/F4c3hugg3r/Go-Chat-Server/pkg/shared"
)

// clients who communicate with the sever
type ChatService struct {
	clients map[string]*Client
	mu      sync.RWMutex
}

func NewChatService() *ChatService {
	return &ChatService{clients: make(map[string]*Client)}
}

// logOutClient deleted a client out of the clients map
func (s *ChatService) logOutClient(clientId string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if client, ok := s.clients[clientId]; ok {
		fmt.Println("logged out ", client.Name)
		close(client.clientCh)
		delete(s.clients, clientId)
		return nil
	} else {
		return fmt.Errorf("client already deleted")
	}

}

// InactiveClientDeleter searches for inactive clients and deletes them as well as closes their message-channel
func (s *ChatService) InactiveClientDeleter() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for clientId, client := range s.clients {
		if !client.Active {
			fmt.Println("due to inactivity: deleting ", client.Name)
			close(client.clientCh)
			delete(s.clients, clientId)
		}
	}
}

// registerClient safely registeres a client by creating a Client with the received values
// and putting it into the global clients map
func (s *ChatService) registerClient(clientId string, body Message) (token string, e error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	token = tokenGenerator.GenerateSecureToken(64)

	if _, exists := s.clients[clientId]; exists {
		return token, fmt.Errorf("client already defined")
	}
	clientCh := make(chan Message)
	s.clients[clientId] = &Client{body.Name, clientId, clientCh, true, token}

	fmt.Printf("\nNew client '%s' registered.\n", body.Name)
	return token, nil
}

// sendBroadcast distributes an incomming message abroad all client channels if
// a client can't receive, i'ts active status is set to false
func (s *ChatService) sendBroadcast(msg Message) {
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
			client.Active = true
		default:
			client.Active = false
		}
	}
}

// echo sends a message to the request submitter
func (s *ChatService) echo(clientId string, msg []string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	content := fmt.Sprint(strings.Join(msg, ", "), "\n")

	if client, ok := s.clients[clientId]; ok {
		client.clientCh <- Message{"Plugin message", content}
	}
}

// getClientChannel tests if there is a registered client to the given clientId ans returns
// it's channel and name
func (s *ChatService) getClient(clientId string) (client *Client, err error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	client, exists := s.clients[clientId]
	if !exists {
		err = fmt.Errorf("there is no client with id: %s registered", clientId)
		return
	}
	err = nil
	return
}

// ListClients returns a string slice containing every client with name
// and active status
func (s *ChatService) ListClients() (clientsSlice []string) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, client := range s.clients {
		jsonString, _ := json.Marshal(client)
		clientsSlice = append(clientsSlice, string(jsonString))
	}
	return
}
