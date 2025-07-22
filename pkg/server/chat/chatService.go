package chat

import (
	"fmt"
	"log"
	"sync"
	"time"

	ty "github.com/F4c3hugg3r/Go-Chat-Server/pkg/shared"
)

// clients who communicate with the sever
type ChatService struct {
	clients  map[string]*Client
	groups   map[string]*Group
	maxUsers int
	mu       sync.RWMutex
}

func NewChatService(maxUsers int) *ChatService {
	return &ChatService{
		clients:  make(map[string]*Client),
		groups:   make(map[string]*Group),
		maxUsers: maxUsers,
	}
}

func (s *ChatService) Broadcast(clientsToIterate map[string]*Client, rsp *ty.Response, clientId string) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	switch clientsToIterate {
	case nil:
		clientsToIterate = s.clients
		for _, client := range clientsToIterate {
			if client.ClientId != clientId && client.GetGroupId() == "" {
				err := client.Send(rsp)
				if err != nil {
					log.Printf("\n%v: %s -> %s", err, rsp.RspName, client.Name)
				}
			}
		}
	default:
		for _, client := range clientsToIterate {
			if client.ClientId != clientId {
				err := client.Send(rsp)
				if err != nil {
					log.Printf("\n%v: %s -> %s", err, rsp.RspName, client.Name)
				}
			}
		}
	}
}

// InactiveObjectDeleter searches for idle clients and deletes them as well as closes their message-channel
func (s *ChatService) InactiveObjectDeleter(timeLimit time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for clientId, client := range s.clients {
		if client.Idle(timeLimit) {
			delete(s.clients, clientId)
		}
	}

	for groupId, group := range s.groups {
		if group.SetSize() < 1 {
			delete(s.groups, groupId)
		}
	}
}

func (s *ChatService) LogOutAllUsers() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, client := range s.clients {
		client.Send(&ty.Response{Content: ty.UnregisterFlag})
	}
}

// Echo sends a response to the request submitter
func (s *ChatService) Echo(clientId string, rsp *ty.Response) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	client, ok := s.clients[clientId]
	if !ok {
		return fmt.Errorf("%w: client not found", ty.ErrNotAvailable)
	}

	return client.Send(rsp)
}

// GetClientChannel tests if there is a registered client to the given clientId and returns it
func (s *ChatService) GetClient(clientId string) (*Client, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	client, exists := s.clients[clientId]
	if !exists {
		return client, fmt.Errorf("%w: there is no client with id: %s registered", ty.ErrNotAvailable, clientId)
	}

	return client, nil
}

func (s *ChatService) GetGroup(groupId string) (*Group, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	group, exists := s.groups[groupId]
	if !exists {
		return group, fmt.Errorf("%w: there is no group with id: %s registered", ty.ErrNotAvailable, groupId)
	}

	return group, nil
}
