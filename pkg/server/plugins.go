package server

import (
	"encoding/json"
	"time"
)

// HelpPlugin tells you information about available plugins
type HelpPlugin struct {
	pr *PluginRegistry
}

func NewHelpPlugin(pr *PluginRegistry) *HelpPlugin {
	return &HelpPlugin{pr: pr}
}

func (h *HelpPlugin) Execute(message *Message) (Response, error) {
	jsonList, err := json.Marshal(ListPlugins(h.pr))
	if err != nil {
		return Response{message.Name, "error parsing plugins to json"}, err
	}
	return Response{message.Name, string(jsonList)}, nil
}

// ListPlugins lists all Plugins with correspontig commands
func ListPlugins(pr *PluginRegistry) []string {
	stringSlice := []string{"You can execute a command by typing one of the the following commands:"}

	for commands := range pr.plugins {
		stringSlice = append(stringSlice, commands)
	}
	return stringSlice
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
		return Response{message.Name, "error parsing users to json"}, err
	}
	return Response{message.Name, string(jsonList)}, nil
}

// TimePlugin tells you the current time
type TimePlugin struct{}

func NewTimePlugin() *TimePlugin {
	return &TimePlugin{}
}

func (t *TimePlugin) Execute(message *Message) (Response, error) {
	return Response{message.Name, time.Now().String()}, nil
}
