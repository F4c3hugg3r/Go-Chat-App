package chat

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	ty "github.com/F4c3hugg3r/Go-Chat-Server/pkg/shared"
)

// CallPlugin forwards webRTC signals (SDP, ICE Candidates) to the other group members
type CallPlugin struct {
	chatService *ChatService
}

func NewCallPlugin(s *ChatService) *CallPlugin {
	return &CallPlugin{chatService: s}
}

func (cp *CallPlugin) Description() *Description {
	return &Description{
		Description: "lets you start a voice call in your group",
		Template:    "/call",
	}
}

func (cp *CallPlugin) Execute(msg *ty.Message) (*ty.Response, error) {
	group, _, err := GetCurrentGroup(msg.ClientId, cp.chatService)
	if err != nil {
		return &ty.Response{Err: fmt.Sprintf("%v: error getting current group", err)}, nil
	}

	if group == nil {
		return &ty.Response{Err: fmt.Sprintf("%v: you are not in a group yet", err)}, nil
	}

	rsp := &ty.Response{ClientId: msg.ClientId, RspName: msg.Name, Content: msg.Content}
	cp.chatService.Broadcast(group.GetClients(), rsp, msg.ClientId)

	return &ty.Response{RspName: ty.StartCallFlag, Content: "idk"}, nil
}

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

func (pp *PrivateMessagePlugin) Execute(msg *ty.Message) (*ty.Response, error) {
	client, err := pp.chatService.GetClient(msg.ClientId)
	if err != nil {
		return &ty.Response{Err: fmt.Sprintf("%v: client with id: %s not found", err, msg.ClientId)}, nil
	}

	rsp := &ty.Response{RspName: fmt.Sprintf("[Private] - %s", msg.Name), Content: msg.Content}

	err = client.Send(rsp)
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

func (lp *LogOutPlugin) Execute(msg *ty.Message) (*ty.Response, error) {
	lp.chatService.mu.Lock()
	defer lp.chatService.mu.Unlock()

	client, ok := lp.chatService.clients[msg.ClientId]
	if !ok {
		return nil, fmt.Errorf("%w: client (probably) already deleted", ty.ErrNotAvailable)
	}

	fmt.Println("\nlogged out ", client.Name)
	client.Close()
	delete(lp.chatService.clients, msg.ClientId)

	go client.Execute(lp.pr, &ty.Message{Name: "", Plugin: "/broadcast", Content: fmt.Sprintf("%s hat den Chat verlassen", msg.Name), ClientId: msg.ClientId})

	return &ty.Response{RspName: msg.Name, Content: "Du hast dich ausgeloggt"}, nil
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

func (rp *RegisterClientPlugin) Execute(msg *ty.Message) (*ty.Response, error) {
	rp.chatService.mu.Lock()
	defer rp.chatService.mu.Unlock()

	if len(rp.chatService.clients) >= rp.chatService.maxUsers {
		return &ty.Response{Err: fmt.Sprintf("%v: usercap %d reached, try again later. users:%d", ty.ErrNoPermission, rp.chatService.maxUsers, len(rp.chatService.clients))}, nil
	}

	if _, exists := rp.chatService.clients[msg.ClientId]; exists {
		return &ty.Response{Err: fmt.Sprintf("%v: client already registered", ty.ErrNoPermission)}, nil
	}

	clientCh := make(chan *ty.Response, 100)
	token := ty.GenerateSecureToken(64)
	client := &Client{
		Name:      msg.Name,
		ClientId:  msg.ClientId,
		clientCh:  clientCh,
		Active:    true,
		authToken: token,
		lastSign:  time.Now().UTC(),
		chClosed:  false,
	}
	rp.chatService.clients[msg.ClientId] = client

	fmt.Printf("\nnew client '%s' registered.", msg.Content)

	go client.Execute(rp.pr, &ty.Message{Name: "", Plugin: "/broadcast", Content: fmt.Sprintf("%s ist dem Chat beigetreten", client.Name), ClientId: client.ClientId})

	return &ty.Response{RspName: msg.Name, Content: token}, nil
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

func (bp *BroadcastPlugin) Execute(msg *ty.Message) (*ty.Response, error) {
	rsp := &ty.Response{RspName: msg.Name, Content: msg.Content}

	if strings.TrimSpace(msg.Content) == "" {
		return rsp, nil
	}

	group, _, err := GetCurrentGroup(msg.Name, bp.chatService)
	if err != nil {
		return &ty.Response{Err: fmt.Sprintf("%v: error getting current group", err)}, nil
	}

	if group != nil {
		bp.chatService.Broadcast(group.GetClients(), rsp, msg.ClientId)
		return rsp, nil
	}

	bp.chatService.Broadcast(nil, rsp, msg.ClientId)
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

func (h *HelpPlugin) Execute(msg *ty.Message) (*ty.Response, error) {
	jsonList, err := json.Marshal(ListPlugins(h.pr.plugins))
	if err != nil {
		return nil, fmt.Errorf("%w: error parsing plugins to json", err)
	}

	return &ty.Response{RspName: "Help", Content: string(jsonList)}, nil
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

func (u *ListUsersPlugin) Execute(msg *ty.Message) (*ty.Response, error) {
	u.chatService.mu.RLock()
	defer u.chatService.mu.RUnlock()

	clientsSlice := GenericMapToJSONSlice(u.chatService.clients)

	jsonList, err := json.Marshal(clientsSlice)
	if err != nil {
		return nil, fmt.Errorf("%w: error parsing clients to json", err)
	}

	return &ty.Response{RspName: "Users", Content: string(jsonList)}, nil
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

func (t *TimePlugin) Execute(msg *ty.Message) (*ty.Response, error) {
	return &ty.Response{RspName: "Time", Content: time.Now().UTC().String()}, nil
}
