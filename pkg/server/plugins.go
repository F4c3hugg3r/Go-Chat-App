package server

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/F4c3hugg3r/Go-Chat-Server/pkg/shared"
)

// PrivateMessage Plugin lets a client send a private message to another client identified by it's clientId
type PrivateMessagePlugin struct {
	chatService *ChatService
}

func NewPrivateMessagePlugin(s *ChatService) *PrivateMessagePlugin {
	return &PrivateMessagePlugin{chatService: s}
}

func (pp *PrivateMessagePlugin) Execute(message *Message) (Response, error) {
	pp.chatService.mu.Lock()
	defer pp.chatService.mu.Unlock()

	client, ok := pp.chatService.clients[message.ClientId]
	if !ok {
		return Response{Name: "client not available"}, fmt.Errorf("%w: client with id: %s not found", ClientNotAvailableError, message.ClientId)
	}

	rsp := Response{fmt.Sprintf("[Private] - %s", message.Name), message.Content}

	select {
	case client.clientCh <- rsp:
		fmt.Printf("successfully sent to %s ", client.Name)

		client.Active = true
		client.lastSign = time.Now()

		return rsp, nil

	case <-time.After(500 * time.Millisecond):
		client.Active = false
		return Response{Name: "client not available"}, fmt.Errorf("%w: private message couldn't be delivered in time", ClientNotAvailableError)
	}
}

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
	if !ok {
		return Response{Name: "client already deleted"}, fmt.Errorf("%w: client (probably) already deleted", ClientNotAvailableError)
	}

	fmt.Println("logged out ", client.Name)
	close(client.clientCh)
	delete(lp.chatService.clients, message.ClientId)

	return Response{Name: message.Name, Content: "logged out"}, nil
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

	if len(rp.chatService.clients) >= rp.chatService.maxUsers {
		return Response{Name: "usercap already reached, try again later"},
			fmt.Errorf("%w: usercap %d reached, try again later. users:%d", NoPermissionError, rp.chatService.maxUsers, len(rp.chatService.clients))
	}

	if _, exists := rp.chatService.clients[message.ClientId]; exists {
		return Response{Name: "client already defined"}, fmt.Errorf("%w: client already defined", NoPermissionError)
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

	if strings.TrimSpace(message.Content) == "" {
		return Response{}, fmt.Errorf("%w: no empty messages allowed", EmptyStringError)
	}

	if len(bp.chatService.clients) <= 0 {
		return Response{"Server: ", "there are no clients registered"}, fmt.Errorf("%w: There are no clients registered", ClientNotAvailableError)
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
		return Response{Name: "error parsing plugins to json"}, fmt.Errorf("%w: error parsing plugins to json", err)
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
		return Response{Name: "error parsing clients to json"}, fmt.Errorf("%w: error parsing clients to json", err)
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
