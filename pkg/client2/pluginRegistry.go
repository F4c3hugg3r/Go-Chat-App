package client2

// import (
// 	"encoding/json"
// 	"log"
// 	"slices"
// )

// type PluginInterface interface {
// 	// if an error accures, response.Content is empty
// 	Execute(message *Message) error
// 	Description() string
// }

// type PluginRegistry struct {
// 	plugins   map[string]PluginInterface
// 	invisible []string
// }

// // RegisterPlugins sets up all the plugins
// func RegisterPlugins(client *ChatClient) *PluginRegistry {
// 	pr := PluginRegistry{plugins: make(map[string]PluginInterface)}
// 	pr.plugins["/help"] = NewHelpPlugin(&pr)
// 	pr.plugins["/time"] = NewTimePlugin()
// 	pr.plugins["/users"] = NewUserPlugin(client)
// 	pr.plugins["/register"] = NewRegisterClientPlugin(client)
// 	pr.plugins["/broadcast"] = NewBroadcastPlugin(client)
// 	pr.plugins["/quit"] = NewLogOutPlugin(client)
// 	pr.plugins["/private"] = NewPrivateMessagePlugin(client)

// 	pr.invisible = append(pr.invisible, "/broadcast")

// 	return &pr
// }

// func (pr *PluginRegistry) FindAndExecute(message *Message) error {
// 	plugin, ok := pr.plugins[message.Plugin]
// 	if !ok {
// 		return nil
// 	}

// 	return plugin.Execute(message)
// }

// // ListPlugins lists all Plugins with correspontig commands
// func (pr *PluginRegistry) ListPlugins() []json.RawMessage {
// 	jsonSlice := []json.RawMessage{}

// 	for command, plugin := range pr.plugins {
// 		if !slices.Contains(pr.invisible, command) {
// 			jsonString, err := json.Marshal(Plugin{Command: command, Description: plugin.Description()})
// 			if err != nil {
// 				log.Printf("error parsing plugin %s to json", command)
// 			}

// 			jsonSlice = append(jsonSlice, jsonString)
// 		}
// 	}

// 	return jsonSlice
// }
