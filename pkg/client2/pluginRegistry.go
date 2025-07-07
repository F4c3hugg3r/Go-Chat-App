package client2

import "fmt"

type PluginInterface interface {
	// if an error accures, response.Content is empty
	Execute(message *Message) func() error
	Description() string
}

type PluginRegistry struct {
	plugins    map[string]PluginInterface
	invisible  []string
	chatClient *ChatClient
}

// RegisterPlugins sets up all the plugins
func RegisterPlugins(chatClient *ChatClient) *PluginRegistry {
	pr := PluginRegistry{plugins: make(map[string]PluginInterface)}
	pr.plugins["/help"] = NewHelpPlugin(chatClient)
	pr.plugins["/time"] = NewTimePlugin(chatClient)
	pr.plugins["/users"] = NewUserPlugin(chatClient)
	pr.plugins["/register"] = NewRegisterClientPlugin(chatClient)
	pr.plugins["/broadcast"] = NewBroadcastPlugin(chatClient)
	pr.plugins["/quit"] = NewLogOutPlugin(chatClient)
	pr.plugins["/private"] = NewPrivateMessagePlugin(chatClient)

	pr.invisible = append(pr.invisible, "/broadcast")
	pr.chatClient = chatClient

	return &pr
}

func (pr *PluginRegistry) FindAndExecute(message *Message) func() error {
	plugin, ok := pr.plugins[message.Plugin]
	if !ok {
		return nil
	}

	registered := pr.chatClient.Registered
	if message.Plugin != "/register" && !registered {
		return func() error {
			return fmt.Errorf("%w: you are not registered yet", ErrNoPermission)
		}
	}
	if message.Plugin == "/register" && registered {
		return func() error {
			return fmt.Errorf("%w: you are already registered", ErrNoPermission)
		}
	}

	return plugin.Execute(message)
}

// // ListPlugins lists all Plugins with correspontig commands
// func (pr *PluginRegistry) ListPlugins() []PluginInterface {
// 	plugins := []PluginInterface{}

// 	for command, plugin := range pr.plugins {
// 		if !slices.Contains(pr.invisible, command) {
// 			plugins = append(plugins, plugin)
// 		}
// 	}

// 	return plugins
// }
