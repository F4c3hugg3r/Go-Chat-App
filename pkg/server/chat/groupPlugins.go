package chat

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"

	ty "github.com/F4c3hugg3r/Go-Chat-Server/pkg/server/types"
	"github.com/F4c3hugg3r/Go-Chat-Server/pkg/shared"
)

// GroupHelpPlugin
type GroupHelpPlugin struct {
	s   *ChatService
	gpr *GroupPluginRegistry
}

func NewGroupHelpPlugin(s *ChatService, gpr *GroupPluginRegistry) *GroupHelpPlugin {
	return &GroupHelpPlugin{
		s:   s,
		gpr: gpr,
	}
}

func (ghp *GroupHelpPlugin) Description() *Description {
	return &Description{
		Description: "tells every group command plus description",
		Template:    "/group help",
	}
}

func (ghp *GroupHelpPlugin) Execute(msg *ty.Message) (*ty.Response, error) {
	jsonList, err := json.Marshal(ListPlugins(ghp.gpr.gPlugins))
	if err != nil {
		return nil, fmt.Errorf("%w: error parsing plugins to json", err)
	}

	return &ty.Response{Name: "Group Help", Content: string(jsonList)}, nil
}

// GroupListPlugin
type GroupListPlugin struct {
	s *ChatService
}

func NewGroupListPlugin(s *ChatService) *GroupListPlugin {
	return &GroupListPlugin{s: s}
}

func (glp *GroupListPlugin) Description() *Description {
	return &Description{
		Description: "lists every group plus info",
		Template:    "/group list",
	}
}

func (glp *GroupListPlugin) Execute(msg *ty.Message) (*ty.Response, error) {
	glp.s.mu.Lock()
	defer glp.s.mu.Unlock()

	if len(glp.s.groups) < 1 {
		return &ty.Response{Err: fmt.Errorf("%v: there are no groups", ty.ErrNotAvailable)}, nil
	}

	groupSlice := []json.RawMessage{}

	for _, group := range glp.s.groups {
		group.SetSize()

		jsonString, err := json.Marshal(group)
		if err != nil {
			log.Printf("error parsing client %s to json", group.Name)
		}

		groupSlice = append(groupSlice, jsonString)
	}

	jsonList, err := json.Marshal(groupSlice)
	if err != nil {
		return nil, fmt.Errorf("%w: error parsing clients to json", err)
	}

	return &ty.Response{Name: "Group List", Content: string(jsonList)}, nil
}

// GroupCreatePlugin
type GroupCreatePlugin struct {
	s *ChatService
}

func NewGroupCreatePlugin(s *ChatService) *GroupCreatePlugin {
	return &GroupCreatePlugin{s: s}
}

func (gcp *GroupCreatePlugin) Description() *Description {
	return &Description{
		Description: "creates a chat group",
		Template:    "/group create {name}",
	}
}

func (gcp *GroupCreatePlugin) Execute(msg *ty.Message) (*ty.Response, error) {
	gcp.s.mu.Lock()
	defer gcp.s.mu.Unlock()

	name := strings.TrimSpace(msg.Content)
	id := shared.GenerateSecureToken(32)
	clients := make(map[string]*Client)

	client, exists := gcp.s.clients[msg.ClientId]
	if !exists {
		return nil, fmt.Errorf("%w: there is no client with id: %s registered", ty.ErrNotAvailable, msg.ClientId)
	}

	clients[msg.ClientId] = client
	group := &Group{GroupId: id, Name: name, clients: clients, mu: &sync.RWMutex{}}
	gcp.s.groups[id] = group
	client.SetGroup(id)

	fmt.Printf("\nnew group %s created", group.Name)

	jsonGroup, err := json.Marshal(group)
	if err != nil {
		return nil, fmt.Errorf("%w: error parsing group to json", err)
	}

	// TODO in client jsonGroup verarbeiten
	return &ty.Response{Name: "Add Group", Content: string(jsonGroup)}, nil
}

// TODO GroupInvitePlugin
// type GroupInvitePlugin struct {
// 	s *ChatService
// }

// func NewGroupInvitePlugin(s *ChatService) *GroupInvitePlugin {
// 	return &GroupInvitePlugin{s: s}
// }

// func (gip *GroupInvitePlugin) Description() *Description {
// 	return &Description{
// 		Description: "invites someone to your group",
// 		Template:    "/group invite {clientId}",
// 	}
// }

// func (gip *GroupInvitePlugin) Execute(msg *ty.Message) (*ty.Response, error) {

// }

// GrouLeavePlugin
type GroupLeavePlugin struct {
	s *ChatService
}

func NewGroupLeavePlugin(s *ChatService) *GroupLeavePlugin {
	return &GroupLeavePlugin{s: s}
}

func (glp *GroupLeavePlugin) Description() *Description {
	return &Description{
		Description: "lets you leave the group",
		Template:    "/group leave",
	}
}

func (glp *GroupLeavePlugin) Execute(msg *ty.Message) (*ty.Response, error) {
	client, err := glp.s.GetClient(msg.ClientId)
	if err != nil {
		return nil, fmt.Errorf("%w: client (probably) already deleted", err)
	}

	group, err := GetCurrentGroup(client, glp.s)
	if err != nil {
		return nil, fmt.Errorf("%w: group (probably) already deleted", err)
	}

	err = group.RemoveClient(client)
	if err != nil {
		return &ty.Response{Err: err}, nil
	}

	client.UnsetGroup()

	return &ty.Response{Name: "Leave Group", Content: "Du hast die Gruppe verlassen"}, nil
}

// GroupUserPlugin
type GroupUsersPlugin struct {
	s *ChatService
}

func NewGroupUsersPlugin(s *ChatService) *GroupUsersPlugin {
	return &GroupUsersPlugin{s: s}
}

func (gup *GroupUsersPlugin) Description() *Description {
	return &Description{
		Description: "shows every user in the group",
		Template:    "/group users",
	}
}

func (gup *GroupUsersPlugin) Execute(msg *ty.Message) (*ty.Response, error) {
	client, err := gup.s.GetClient(msg.ClientId)
	if err != nil {
		return nil, fmt.Errorf("%w: client (probably) already deleted", ty.ErrNotAvailable)
	}

	group, err := GetCurrentGroup(client, gup.s)
	if err != nil {
		return &ty.Response{Err: fmt.Errorf("%w: error finding group", err)}, nil
	}

	if group == nil {
		return &ty.Response{Err: fmt.Errorf("%w: you are not in a group", ty.ErrNoPermission)}, nil
	}

	groupsSlice := group.SafeGetGroupSlice()

	jsonList, err := json.Marshal(groupsSlice)
	if err != nil {
		return nil, fmt.Errorf("%w: error parsing groups to json", err)
	}

	return &ty.Response{Name: "Group Users", Content: string(jsonList)}, nil
}

// GroupJoinPlugin
type GroupJoinPlugin struct {
	s *ChatService
}

func NewGroupJoinPlugin(s *ChatService) *GroupJoinPlugin {
	return &GroupJoinPlugin{s: s}
}

func (gjp *GroupJoinPlugin) Description() *Description {
	return &Description{
		Description: "lets you join a group",
		Template:    "/group join {groupId}",
	}
}

func (gjp *GroupJoinPlugin) Execute(msg *ty.Message) (*ty.Response, error) {
	groupId := strings.TrimSpace(msg.Content)

	client, err := gjp.s.GetClient(msg.ClientId)
	if err != nil {
		return nil, fmt.Errorf("%w: client (probably) already deleted", ty.ErrNotAvailable)
	}

	currentGroup := client.GetGroupId()

	group, err := gjp.s.GetGroup(groupId)
	if err != nil {
		return &ty.Response{Err: fmt.Errorf("%w: error finding group with id %s", err, groupId)}, nil
	}

	if currentGroup != "" {
		client.UnsetGroup()
		group.RemoveClient(client)
	}

	err = group.AddClient(client)
	if err != nil {
		return &ty.Response{Err: err}, nil
	}

	client.SetGroup(groupId)

	return &ty.Response{Name: "Add Group", Content: "Du bist der Gruppe beigetreten"}, nil
}
