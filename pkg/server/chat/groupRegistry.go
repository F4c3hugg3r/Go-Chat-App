package chat

import (
	"fmt"
	"sync"

	ty "github.com/F4c3hugg3r/Go-Chat-Server/pkg/server/types"
)

type Group struct {
	clients map[string]*Client
	Id      string
	Name    string
	mu      *sync.RWMutex
	Size    int
}

type GroupPluginRegistry struct {
	gPlugins map[string]PluginInterface
}

// TODO autodelete
func RegisterGroupPlugins(s *ChatService) *GroupPluginRegistry {
	gp := &GroupPluginRegistry{gPlugins: make(map[string]PluginInterface)}
	gp.gPlugins["help"] = NewGroupHelpPlugin(s, gp)
	gp.gPlugins["list"] = NewGroupListPlugin(s)
	gp.gPlugins["join"] = NewGroupJoinPlugin(s)
	gp.gPlugins["create"] = NewGroupCreatePlugin(s)
	gp.gPlugins["leave"] = NewGroupLeavePlugin(s)
	gp.gPlugins["users"] = NewGroupUsersPlugin(s)
	// gp.gPlugins["invite"] = NewGroupInvitePlugin(s)
	// rules

	// Für private:
	// kick
	// admin

	return gp
}

func (gp *GroupPluginRegistry) Description() *Description {
	return gp.gPlugins["help"].Description()
}

func (gp *GroupPluginRegistry) Execute(message *ty.Message) (*ty.Response, error) {
	newMsg, err := extractIdentifierMessage(message)
	if err != nil {
		return nil, fmt.Errorf("%w: message couldn't be parsed to group command", err)
	}

	plugin, ok := gp.gPlugins[newMsg.Plugin]
	if !ok {
		return &ty.Response{Err: fmt.Errorf("%w: no such group command identifier found: %s", ty.ErrNoPermission, newMsg.Plugin)}, nil
	}

	return plugin.Execute(message)
}

func (g *Group) GetClients() map[string]*Client {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return g.clients
}

func (g *Group) SetSize() {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.Size = len(g.clients)
}

func (g *Group) AddClient(client *Client) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if _, exists := g.clients[client.ClientId]; !exists && client != nil {
		g.clients[client.ClientId] = client
		return nil
	}

	return fmt.Errorf("%w: you are already in this group", ty.ErrNoPermission)
}

func (g *Group) RemoveClient(client *Client) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if _, exists := g.clients[client.ClientId]; exists && client != nil {
		delete(g.clients, client.ClientId)
		return nil
	}

	return fmt.Errorf("%w: you are not in this group", ty.ErrNoPermission)
}
