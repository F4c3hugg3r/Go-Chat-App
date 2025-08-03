package chat

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	ty "github.com/F4c3hugg3r/Go-Chat-Server/pkg/shared"
)

// CallPlugin returns a slice of all the other group member ids
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
		return &ty.Response{Err: fmt.Sprintf("%v: you are not in a group yet", ty.ErrNoPermission)}, nil
	}

	if group.SetSize() > 6 {
		return &ty.Response{Err: fmt.Sprintf("%v: group is to big for a call (max 6 clients)", ty.ErrNoPermission)}, nil
	}

	groupClientIds := group.ConnectToGroupMembers(msg.ClientId)

	jsonSlice := json.RawMessage{}
	jsonSlice, err = json.Marshal(groupClientIds)
	if err != nil {
		return nil, fmt.Errorf("%w: error encoding clientId to json", err)
	}

	return &ty.Response{RspName: msg.Name, Content: string(jsonSlice), Err: ty.IgnoreResponseTag}, nil
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

	if client.groupId != "" {
		delete(lp.chatService.groups[client.groupId].clients, client.ClientId)
	}

	fmt.Printf("\nlogged out %s", client.Name)
	client.Close()
	delete(lp.chatService.clients, client.ClientId)

	go lp.chatService.Broadcast(nil, &ty.Response{RspName: ty.UserRemoveFlag, Content: msg.Name, ClientId: msg.ClientId})

	return &ty.Response{RspName: msg.Name, Content: ty.UnregisterFlag}, nil
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
		GroupName: "",
		clientCh:  clientCh,
		active:    true,
		authToken: token,
		lastSign:  time.Now().UTC(),
		chClosed:  false,
	}
	rp.chatService.clients[msg.ClientId] = client

	fmt.Printf("\nnew client '%s' registered.", msg.Content)

	go rp.chatService.Broadcast(nil, &ty.Response{RspName: ty.UserAddFlag, Content: client.Name, ClientId: msg.ClientId})

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
	rsp := &ty.Response{RspName: msg.Name, Content: msg.Content, ClientId: msg.ClientId}

	if strings.TrimSpace(msg.Content) == "" {
		return rsp, nil
	}

	group, _, err := GetCurrentGroup(msg.ClientId, bp.chatService)
	if err != nil {
		return &ty.Response{Err: fmt.Sprintf("%v: error getting current group", err)}, nil
	}

	if group != nil {
		bp.chatService.Broadcast(group.GetClients(), rsp)
		return rsp, nil
	}

	bp.chatService.Broadcast(nil, rsp)
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

	clientsSlice := ClientsToJsonSliceRequireLock(u.chatService.clients, msg.ClientId)

	jsonList, err := json.Marshal(clientsSlice)
	if err != nil {
		return nil, fmt.Errorf("%w: error parsing clients to json", err)
	}

	return &ty.Response{RspName: ty.UsersFlag, Content: string(jsonList)}, nil
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
