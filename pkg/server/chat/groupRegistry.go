package chat

import (
	"encoding/json"
	"fmt"
	"slices"
	"sort"
	"strings"
	"sync"

	ty "github.com/F4c3hugg3r/Go-Chat-Server/pkg/shared"
)

type Group struct {
	clients map[string]*Client
	// key: composite key from both clientIds, value: ICE Connected
	rtcs    map[string]bool
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

	// TODO
	// kick
	// admin
	// invite
	// rules

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

// GetClientIdsFromGroup return every clientId which is not in a rtc with given clientId
func (g *Group) GetClientIdsFromGroup(ownId string, onlyCallable bool) []string {
	g.mu.RLock()
	defer g.mu.RUnlock()

	var inCallOppKeys []string

	if onlyCallable {
		for compKey := range g.rtcs {
			if strings.Contains(compKey, ownId) {
				inCallOppKeys = append(inCallOppKeys, GetRTCPartnerFromKey(ownId, compKey))
			}
		}
	}

	var clientIds []string
	for clientId, _ := range g.clients {
		if clientId != ownId {
			if slices.Contains(inCallOppKeys, clientId) && onlyCallable {
				continue
			}
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

// CheckConnections returns every current rtc (doesn't have to be ICE connected)
// of one user and returns the composite key
// func (g *Group) CheckConnections(clientId string) []string {
// 	g.mu.RLock()
// 	defer g.mu.RUnlock()

// 	var compKeysSlice []string

// 	for compKey := range g.connections {
// 		if strings.Contains(compKey, clientId) {
// 			compKeysSlice = append(compKeysSlice, compKey)
// 		}
// 	}
// 	return compKeysSlice
// }

func (g *Group) ConnectToGroupMembers(ownId string) []string {
	stringSlice := g.GetClientIdsFromGroup(ownId, true)

	for _, oppId := range stringSlice {
		g.AddConnection(ownId, oppId)
	}

	return stringSlice
}

func (g *Group) AddConnection(ownId string, oppId string) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.rtcs[CreateCompositeKey(ownId, oppId)] = false
}

func CreateCompositeKey(firstId string, secondId string) string {
	ids := []string{firstId, secondId}
	sort.Strings(ids)
	return ids[0] + ":" + ids[1]
}

func GetRTCPartnerFromKey(ownId string, compKey string) string {
	oppId := strings.Replace(compKey, ownId, "", -1)
	return strings.Replace(oppId, ":", "", -1)
}
