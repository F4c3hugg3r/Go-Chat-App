package client2

import "fmt"

const (
	UnregisteredOnly = iota
	RegisteredOnly
	Always
)

type PluginInterface interface {
	// if an error accures, response.Content is empty
	Execute(message *Message) func() error
	Description() string
	CheckScope() int
}

type PluginRegistry struct {
	plugins           map[string]PluginInterface
	registrationScope map[string]int
	chatClient        *ChatClient
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

	pr.chatClient = chatClient

	return &pr
}

func (pr *PluginRegistry) FindAndExecute(message *Message) func() error {
	return func() error {
		plugin, ok := pr.plugins[message.Plugin]
		if !ok {
			return fmt.Errorf("%w: plugin not found", ErrNoPermission)
		}

		scope := pr.plugins[message.Plugin].CheckScope()

		switch scope {
		case UnregisteredOnly:
			if pr.chatClient.Registered {
				return fmt.Errorf("%w: you are already registered", ErrNoPermission)
			}
		case RegisteredOnly:
			if !pr.chatClient.Registered {
				return fmt.Errorf("%w: you are not registered yet", ErrNoPermission)
			}
		}
		return plugin.Execute(message)()
	}
}
