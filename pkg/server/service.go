package server

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"
)

// clients who communicate with the sever
type ChatService struct {
	clients  map[string]*Client
	maxUsers int
	mu       sync.RWMutex
}

func NewChatService(maxUsers int) *ChatService {
	return &ChatService{clients: make(map[string]*Client), maxUsers: maxUsers}
}

// InactiveClientDeleter searches for inactive clients and deletes them as well as closes their message-channel
func (s *ChatService) InactiveClientDeleter(timeLimit time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for clientId, client := range s.clients {
		if time.Since(client.lastSign) >= timeLimit {
			client.Active = false
		}

		if !client.Active {
			fmt.Println("due to inactivity: deleting ", client.Name)
			select {
			case client.clientCh <- Response{Name: "inactive"}:
			case <-time.After(500 * time.Millisecond):
			}
			close(client.clientCh)
			delete(s.clients, clientId)
		}
	}
}

// echo sends a response to the request submitter
func (s *ChatService) echo(clientId string, rsp Response) {
	s.mu.Lock()
	defer s.mu.Unlock()

	client, ok := s.clients[clientId]
	if ok {
		select {
		case client.clientCh <- rsp:
			fmt.Println("success")

			client.Active = true
			client.lastSign = time.Now()
		case <-time.After(500 * time.Millisecond):
			client.Active = false
		}
	}
}

// getClientChannel tests if there is a registered client to the given clientId and returns
// it's channel and name
func (s *ChatService) getClient(clientId string) (*Client, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	client, exists := s.clients[clientId]
	if !exists {
		return client, fmt.Errorf("there is no client with id: %s registered", clientId)
	}

	return client, nil
}

// ListClients returns a string slice containing every client with name
// and active status
func (s *ChatService) ListClients() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	clientsSlice := []string{}

	for _, client := range s.clients {
		clientsSlice = append(clientsSlice, fmt.Sprintf("Name: %s, Active: %t, Id; %s\n", client.Name, client.Active, client.ClientId))
	}

	return clientsSlice
}

// DecodeToMessage decodes a responseBody to a Message struct
func DecodeToMessage(body []byte) (Message, error) {
	message := Message{}
	dec := json.NewDecoder(strings.NewReader(string(body)))
	err := dec.Decode(&message)

	if err != nil {
		return message, err
	}

	return message, nil
}
