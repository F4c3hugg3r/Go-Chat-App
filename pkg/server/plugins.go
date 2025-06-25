package server

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/F4c3hugg3r/Go-Chat-Server/pkg/shared"
)

// LogOutPlugin logs out a client by deleting it out of the clients map
type LogOutPlugin struct {
	chatService *ChatService
}

func NewLogOutPlugin(s *ChatService) *LogOutPlugin {
	return &LogOutPlugin{chatService: s}
}

func (lp *LogOutPlugin) Execute(message *Message) (Response, error) {
	lp.chatService.mu.Lock()
	defer lp.chatService.mu.Unlock()

	client, ok := lp.chatService.clients[message.ClientId]
	if ok {
		fmt.Println("logged out ", client.Name)
		close(client.clientCh)
		delete(lp.chatService.clients, message.ClientId)
		return Response{Name: message.Name, Content: "logged out"}, nil
	}
	return Response{Name: "client already deleted"}, fmt.Errorf("client already deleted")
}

// RegisterClientPlugin safely registeres a client by creating a Client with the received values
// and putting it into the global clients map
type RegisterClientPlugin struct {
	chatService *ChatService
}

func NewRegisterClientPlugin(s *ChatService) *RegisterClientPlugin {
	return &RegisterClientPlugin{chatService: s}
}

func (rp *RegisterClientPlugin) Execute(message *Message) (Response, error) {
	rp.chatService.mu.Lock()
	defer rp.chatService.mu.Unlock()

	if _, exists := rp.chatService.clients[message.ClientId]; exists {
		return Response{Name: "client already defined"}, fmt.Errorf("client already defined")
	}

	clientCh := make(chan Response, 100)
	token := shared.GenerateSecureToken(64)
	rp.chatService.clients[message.ClientId] = &Client{message.Content, message.ClientId, clientCh, true, token, time.Now()}

	fmt.Printf("\nNew client '%s' registered.\n", message.Content)
	return Response{Name: "authToken", Content: token}, nil
}

// BroadcaastPlugin distributes an incomming message abroad all client channels if
// a client can't receive, i'ts active status is set to false
type BroadcastPlugin struct {
	chatService *ChatService
}

func NewBroadcastPlugin(s *ChatService) *BroadcastPlugin {
	return &BroadcastPlugin{chatService: s}
}

func (bp *BroadcastPlugin) Execute(message *Message) (Response, error) {
	bp.chatService.mu.Lock()
	defer bp.chatService.mu.Unlock()

	if len(bp.chatService.clients) <= 0 {
		return Response{message.Name, "There are no clients registered"}, fmt.Errorf("There are no clients registered")
	}

	rsp := Response{Name: message.Name, Content: message.Content}

	for _, client := range bp.chatService.clients {
		if client.ClientId != message.ClientId {
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
	return rsp, nil
}

// HelpPlugin tells you information about available plugins
type HelpPlugin struct {
	pr *PluginRegistry
}

func NewHelpPlugin(pr *PluginRegistry) *HelpPlugin {
	return &HelpPlugin{pr: pr}
}

func (h *HelpPlugin) Execute(message *Message) (Response, error) {
	jsonList, err := json.Marshal(h.pr.ListPlugins())
	if err != nil {
		return Response{Name: "error parsing plugins to json"}, err
	}
	return Response{"Help", string(jsonList)}, nil
}

// UserPlugin tells you information about all the current users
type UserPlugin struct {
	chatService *ChatService
}

func NewUserPlugin(s *ChatService) *UserPlugin {
	return &UserPlugin{chatService: s}
}

func (u *UserPlugin) Execute(message *Message) (Response, error) {
	jsonList, err := json.Marshal(u.chatService.ListClients())
	if err != nil {
		return Response{Name: "error parsing users to json"}, err
	}
	return Response{"Users", string(jsonList)}, nil
}

// TimePlugin tells you the current time
type TimePlugin struct{}

func NewTimePlugin() *TimePlugin {
	return &TimePlugin{}
}

func (t *TimePlugin) Execute(message *Message) (Response, error) {
	return Response{"Time", time.Now().String()}, nil
}
