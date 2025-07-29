package chat

import (
	"encoding/json"
	"fmt"
	"sync"

	ty "github.com/F4c3hugg3r/Go-Chat-Server/pkg/shared"
)

type Group struct {
	clients map[string]*Client
	GroupId string `json:"groupId"`
	Name    string `json:"name"`
	mu      *sync.RWMutex
	Size    int `json:"size"`
}

type GroupPluginRegistry struct {
	gPlugins map[string]PluginInterface
}

func RegisterGroupPlugins(s *ChatService, pr *PluginRegistry) *GroupPluginRegistry {
	gp := &GroupPluginRegistry{gPlugins: make(map[string]PluginInterface)}
	gp.gPlugins["help"] = NewGroupHelpPlugin(s, gp)
	gp.gPlugins["list"] = NewGroupListPlugin(s)
	gp.gPlugins["join"] = NewGroupJoinPlugin(s, pr)
	gp.gPlugins["create"] = NewGroupCreatePlugin(s)
	gp.gPlugins["leave"] = NewGroupLeavePlugin(s, pr)
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
		return &ty.Response{Err: fmt.Sprintf("%v: no empty identifier allowed", err)}, nil
	}

	plugin, ok := gp.gPlugins[newMsg.Plugin]
	if !ok {
		return &ty.Response{Err: fmt.Sprintf("%v: no such group command identifier found: %s", ty.ErrNoPermission, newMsg.Plugin)}, nil
	}

	return plugin.Execute(message)
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

func (g *Group) GetClientIdsFromGroup(notIncludedId string) []string {
	g.mu.RLock()
	defer g.mu.RUnlock()

	var clientIds []string
	for clientId := range g.clients {
		if clientId != notIncludedId {
			clientIds = append(clientIds, clientId)
		}
	}

	return clientIds
}

func (g *Group) GetClients() map[string]*Client {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return g.clients
}

func (g *Group) SetSize() int {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.Size = len(g.clients)
	return g.Size
}

func (g *Group) SafeGetGroupSlice() []json.RawMessage {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return GenericMapToJSONSlice(g.clients)
}
