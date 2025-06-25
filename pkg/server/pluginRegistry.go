package server

import (
	"fmt"
	"slices"
)

type PluginInterface interface {
	// if an error accures, respponse.Content is empty
	Execute(message *Message) (Response, error)
}

type PluginRegistry struct {
	plugins   map[string]PluginInterface
	invisible []string
}

func RegisterPlugins(chatService *ChatService) *PluginRegistry {
	pr := PluginRegistry{plugins: make(map[string]PluginInterface)}
	pr.plugins["/help"] = NewHelpPlugin(&pr)
	pr.plugins["/time"] = NewTimePlugin()
	pr.plugins["/users"] = NewUserPlugin(chatService)
	pr.plugins["/register"] = NewRegisterClientPlugin(chatService)
	pr.plugins["/broadcast"] = NewBroadcastPlugin(chatService)
	pr.plugins["/quit"] = NewLogOutPlugin(chatService)
	//	pr.plugins["/private"] = NewPrivateMessagePlugin(chatService)

	pr.invisible = append(pr.invisible, "/register", "/broadcast")
	return &pr
}

func (pr *PluginRegistry) FindAndExecute(message *Message) (Response, error) {
	if plugin, ok := pr.plugins[message.Plugin]; ok {
		return plugin.Execute(message)
	}
	return Response{message.Name, fmt.Sprintf("no such plugin found: %s", message.Plugin)}, fmt.Errorf("no such plugin found: %s", message.Plugin)
}

// ListPlugins lists all Plugins with correspontig commands
func (pr *PluginRegistry) ListPlugins() []string {
	stringSlice := []string{"You can execute a command by typing one of the the following commands:"}

	for command := range pr.plugins {
		if !slices.Contains(pr.invisible, command) {
			stringSlice = append(stringSlice, command)
		}
	}
	return stringSlice
}

// TODO Vorschlag
// /help abstrahieren, sodass man /help für jedes Plugin aufrufen kann
// dafür Help() Funktion im Interface
