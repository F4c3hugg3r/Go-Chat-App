package server

import (
	"encoding/json"
	"fmt"
	"log"
)

type PluginInterface interface {
	// if an error accures, respponse.Content is empty
	Execute(message *Message) (Response, error)
}

type PluginRegistry struct {
	plugins   map[string]PluginInterface
	invisible []string
}

// RegisterPlugins sets up all the plugins
func RegisterPlugins(chatService *ChatService) *PluginRegistry {
	pr := PluginRegistry{plugins: make(map[string]PluginInterface)}
	pr.plugins["/help"] = NewHelpPlugin(&pr)
	pr.plugins["/time"] = NewTimePlugin()
	pr.plugins["/users"] = NewUserPlugin(chatService)
	pr.plugins["/register"] = NewRegisterClientPlugin(chatService)
	pr.plugins["/broadcast"] = NewBroadcastPlugin(chatService)
	pr.plugins["/quit"] = NewLogOutPlugin(chatService)
	pr.plugins["/private"] = NewPrivateMessagePlugin(chatService)

	pr.invisible = append(pr.invisible, "/register", "/broadcast")

	return &pr
}

func (pr *PluginRegistry) FindAndExecute(message *Message) (Response, error) {
	plugin, ok := pr.plugins[message.Plugin]
	if !ok {
		return Response{message.Name, fmt.Sprintf("no such plugin found: %s", message.Plugin)}, fmt.Errorf("no such plugin found: %s", message.Plugin)
	}

	return plugin.Execute(message)
}

// ListPlugins lists all Plugins with correspontig commands
func (pr *PluginRegistry) ListPlugins() []json.RawMessage {
	pluginSlice := []Plugin{
		{Command: "/help", Description: "tells every plugin and their description"},
		{Command: "/time", Description: "tells you the current time"},
		{Command: "/users", Description: "tells you information about all the current users"},
		{Command: "/private", Description: "lets you send a private message to someone - template: '/private {Id} {message}'"},
		{Command: "/quit", Description: "loggs you out of the chat"},
	}

	jsonSlice := []json.RawMessage{}

	for _, plugin := range pluginSlice {
		jsonString, err := json.Marshal(plugin)
		if err != nil {
			log.Printf("error parsing plugin %s to json", plugin.Command)
		}

		jsonSlice = append(jsonSlice, jsonString)
	}

	return jsonSlice
}
