package chat

import (
	"encoding/json"
	"fmt"
	"log"

	ty "github.com/F4c3hugg3r/Go-Chat-Server/pkg/server/types"
)

type PluginInterface interface {
	// if an error accures, respponse.Content is empty
	Execute(message *ty.Message) (*ty.Response, error)
	// should display the regex/template in which the command should be typed in
	Description() string
}

type PluginRegistry struct {
	plugins map[string]PluginInterface
}

type Plugin struct {
	Command     string
	Description string
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

func (pr *PluginRegistry) FindAndExecute(message *ty.Message) (*ty.Response, error) {
	plugin, ok := pr.plugins[message.Plugin]
	if !ok {
		return &ty.Response{Name: "Server", Content: fmt.Sprintf("no such plugin found: %s", message.Plugin)}, nil
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
