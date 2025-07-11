package server

import (
	"encoding/json"
	"fmt"
	"log"
)

type PluginInterface interface {
	// if an error accures, respponse.Content is empty
	Execute(message *Message) (*Response, error)
	// should display the regex/template in which the command should be typed in
	Description() string
}

type PluginRegistry struct {
	plugins map[string]PluginInterface
}

// RegisterPlugins sets up all the plugins
func RegisterPlugins(chatService *ChatService) *PluginRegistry {
	pr := &PluginRegistry{plugins: make(map[string]PluginInterface)}
	pr.plugins["/help"] = NewHelpPlugin(pr)
	pr.plugins["/time"] = NewTimePlugin()
	pr.plugins["/users"] = NewUserPlugin(chatService)
	pr.plugins["/register"] = NewRegisterClientPlugin(chatService, pr)
	pr.plugins["/broadcast"] = NewBroadcastPlugin(chatService)
	pr.plugins["/quit"] = NewLogOutPlugin(chatService, pr)
	pr.plugins["/private"] = NewPrivateMessagePlugin(chatService)

	return pr
}

func (pr *PluginRegistry) FindAndExecute(message *Message) (*Response, error) {
	plugin, ok := pr.plugins[message.Plugin]
	if !ok {
		return &Response{"Server", fmt.Sprintf("no such plugin found: %s", message.Plugin)}, nil
	}

	return plugin.Execute(message)
}

// ListPlugins lists all Plugins with correspontig commands
func (pr *PluginRegistry) ListPlugins() []json.RawMessage {
	jsonSlice := []json.RawMessage{}

	for command, plugin := range pr.plugins {
		jsonString, err := json.Marshal(Plugin{Command: command, Description: plugin.Description()})
		if err != nil {
			log.Printf("error parsing plugin %s to json", command)
		}

		jsonSlice = append(jsonSlice, jsonString)
	}

	return jsonSlice
}
