package chat

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	ty "github.com/F4c3hugg3r/Go-Chat-Server/pkg/server/types"
)

// ListPlugins lists all Plugins with correspontig commands
func ListPlugins(plugins map[string]PluginInterface) []json.RawMessage {
	jsonSlice := []json.RawMessage{}

	for command, plugin := range plugins {
		desc := plugin.Description()
		jsonString, err := json.Marshal(Plugin{Command: desc.Template, Description: desc.Description})
		if err != nil {
			log.Printf("error parsing plugin %s to json", command)
			continue
		}

		jsonSlice = append(jsonSlice, jsonString)
	}

	return jsonSlice
}

func GetCurrentGroup(client *Client, s *ChatService) (*Group, error) {
	groupId := client.GetGroupId()
	if groupId == "" {
		return nil, nil
	}

	group, err := s.GetGroup(groupId)
	if err != nil {
		return nil, fmt.Errorf("%w: group not found", err)
	}

	return group, nil
}

func GenericMapToJSONSlice[T any](items map[string]T) []json.RawMessage {
	var result []json.RawMessage

	for _, item := range items {
		jsonBytes, err := json.Marshal(item)
		if err != nil {
			log.Printf("error parsing item %T to json", item)
			continue
		}
		result = append(result, jsonBytes)
	}

	return result
}

func extractIdentifierMessage(msg *ty.Message) (*ty.Message, error) {
	identifier := strings.Fields(msg.Content)[0]
	content, _ := strings.CutPrefix(msg.Content, fmt.Sprintf("%s ", identifier))
	msg.Plugin = identifier
	msg.Content = content

	return msg, nil
}
