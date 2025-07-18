package chat

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	ty "github.com/F4c3hugg3r/Go-Chat-Server/pkg/server/types"
	"github.com/F4c3hugg3r/Go-Chat-Server/pkg/shared"
)

// PrivateMessage Plugin lets a client send a private message to another client identified by it's clientId
type PrivateMessagePlugin struct {
	chatService *ChatService
}

func NewPrivateMessagePlugin(s *ChatService) *PrivateMessagePlugin {
	return &PrivateMessagePlugin{chatService: s}
}

func (pp *PrivateMessagePlugin) Description() *Description {
	return &Description{
		Description: "lets you send a private message",
		Template:    "/private {Id} {message}",
	}
}

func (pp *PrivateMessagePlugin) Execute(message *ty.Message) (*ty.Response, error) {
	pp.chatService.mu.RLock()
	defer pp.chatService.mu.RUnlock()

	client, ok := pp.chatService.clients[message.ClientId]
	if !ok {
		return &ty.Response{Err: fmt.Sprintf("%v: client with id: %s not found", ty.ErrNotAvailable, message.ClientId)}, nil
	}

	rsp := &ty.Response{Name: fmt.Sprintf("[Private] - %s", message.Name), Content: message.Content}

	err := client.Send(rsp)
	if err != nil {
		return nil, err
	}

	return rsp, nil
}

// LogOutPlugin logs out a client by deleting it out of the clients map
type LogOutPlugin struct {
	chatService *ChatService
	pr          *PluginRegistry
}

func NewLogOutPlugin(s *ChatService, pr *PluginRegistry) *LogOutPlugin {
	return &LogOutPlugin{
		chatService: s,
		pr:          pr,
	}
}

func (lp *LogOutPlugin) Description() *Description {
	return &Description{
		Description: "loggs you out of the chat",
		Template:    "/quit",
	}
}

func (lp *LogOutPlugin) Execute(message *ty.Message) (*ty.Response, error) {
	lp.chatService.mu.Lock()
	defer lp.chatService.mu.Unlock()

	client, ok := lp.chatService.clients[message.ClientId]
	if !ok {
		return nil, fmt.Errorf("%w: client (probably) already deleted", ty.ErrNotAvailable)
	}

	fmt.Println("\nlogged out ", client.Name)
	client.Close()
	delete(lp.chatService.clients, message.ClientId)

	go client.Execute(lp.pr, &ty.Message{Name: "", Plugin: "/broadcast", Content: fmt.Sprintf("%s hat den Chat verlassen", message.Name), ClientId: message.ClientId})

	return &ty.Response{Name: message.Name, Content: "Du hast dich ausgeloggt"}, nil
}

// RegisterClientPlugin safely registeres a client by creating a Client with the received values
// and putting it into the global clients map
type RegisterClientPlugin struct {
	chatService *ChatService
	pr          *PluginRegistry
}

func NewRegisterClientPlugin(s *ChatService, pr *PluginRegistry) *RegisterClientPlugin {
	return &RegisterClientPlugin{
		chatService: s,
		pr:          pr,
	}
}

func (rp *RegisterClientPlugin) Description() *Description {
	return &Description{
		Description: "registeres a client",
		Template:    "/register {name}",
	}
}

func (rp *RegisterClientPlugin) Execute(message *ty.Message) (*ty.Response, error) {
	rp.chatService.mu.Lock()
	defer rp.chatService.mu.Unlock()

	if len(rp.chatService.clients) >= rp.chatService.maxUsers {
		return &ty.Response{Err: fmt.Sprintf("%v: usercap %d reached, try again later. users:%d", ty.ErrNoPermission, rp.chatService.maxUsers, len(rp.chatService.clients))}, nil
	}

	if _, exists := rp.chatService.clients[message.ClientId]; exists {
		return &ty.Response{Err: fmt.Sprintf("%v: client already registered", ty.ErrNoPermission)}, nil
	}

	clientCh := make(chan *ty.Response, 100)
	token := shared.GenerateSecureToken(64)
	client := &Client{
		Name:      message.Name,
		ClientId:  message.ClientId,
		clientCh:  clientCh,
		Active:    true,
		authToken: token,
		lastSign:  time.Now().UTC(),
		chClosed:  false,
	}
	rp.chatService.clients[message.ClientId] = client

	fmt.Printf("\nnew client '%s' registered.", message.Content)

	go client.Execute(rp.pr, &ty.Message{Name: "", Plugin: "/broadcast", Content: fmt.Sprintf("%s ist dem Chat beigetreten", client.Name), ClientId: client.ClientId})

	return &ty.Response{Name: message.Name, Content: token}, nil
}

// BroadcaastPlugin distributes an incomming message abroad all client channels if
// a client can't receive, i'ts active status is set to false
type BroadcastPlugin struct {
	chatService *ChatService
}

func NewBroadcastPlugin(s *ChatService) *BroadcastPlugin {
	return &BroadcastPlugin{chatService: s}
}

func (bp *BroadcastPlugin) Description() *Description {
	return &Description{
		Description: "distributes a message abroad all clients",
		Template:    "{message} | /broadcast {message}",
	}
}

func (bp *BroadcastPlugin) Execute(message *ty.Message) (*ty.Response, error) {
	rsp := &ty.Response{Name: message.Name, Content: message.Content}

	if strings.TrimSpace(message.Content) == "" {
		return rsp, nil
	}

	client, err := bp.chatService.GetClient(message.ClientId)
	if err != nil {
		return nil, fmt.Errorf("%w: client (probably) already deleted", err)
	}

	group, err := GetCurrentGroup(client, bp.chatService)
	if err != nil {
		return nil, fmt.Errorf("%w: error finding group", err)
	}

	if group != nil {
		bp.chatService.Broadcast(group.GetClients(), rsp, message.ClientId)
		return rsp, nil
	}

	bp.chatService.Broadcast(nil, rsp, message.ClientId)
	return rsp, nil
}

// HelpPlugin tells you information about available plugins
type HelpPlugin struct {
	pr *PluginRegistry
}

func NewHelpPlugin(pr *PluginRegistry) *HelpPlugin {
	return &HelpPlugin{pr: pr}
}

func (h *HelpPlugin) Description() *Description {
	return &Description{
		Description: "tells every plugin plus description",
		Template:    "/help",
	}
}

func (h *HelpPlugin) Execute(message *ty.Message) (*ty.Response, error) {
	jsonList, err := json.Marshal(ListPlugins(h.pr.plugins))
	if err != nil {
		return nil, fmt.Errorf("%w: error parsing plugins to json", err)
	}

	return &ty.Response{Name: "Help", Content: string(jsonList)}, nil
}

// ListUsersPlugin tells you information about all the current users
type ListUsersPlugin struct {
	chatService *ChatService
}

func NewListUsersPlugin(s *ChatService) *ListUsersPlugin {
	return &ListUsersPlugin{chatService: s}
}

func (u *ListUsersPlugin) Description() *Description {
	return &Description{
		Description: "tells information about all current users",
		Template:    "/users",
	}
}

func (u *ListUsersPlugin) Execute(message *ty.Message) (*ty.Response, error) {
	u.chatService.mu.RLock()
	defer u.chatService.mu.RUnlock()

	clientsSlice := GenericMapToJSONSlice(u.chatService.clients)

	jsonList, err := json.Marshal(clientsSlice)
	if err != nil {
		return nil, fmt.Errorf("%w: error parsing clients to json", err)
	}

	return &ty.Response{Name: "Users", Content: string(jsonList)}, nil
}

// TimePlugin tells you the current time
type TimePlugin struct{}

func NewTimePlugin() *TimePlugin {
	return &TimePlugin{}
}

func (t *TimePlugin) Description() *Description {
	return &Description{
		Description: "tells you the current time",
		Template:    "/time",
	}
}

func (t *TimePlugin) Execute(message *ty.Message) (*ty.Response, error) {
	return &ty.Response{Name: "Time", Content: time.Now().UTC().String()}, nil
}
