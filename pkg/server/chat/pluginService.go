package chat

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	ty "github.com/F4c3hugg3r/Go-Chat-Server/pkg/shared"
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

func GroupClientIdsToJson(group *Group) (json.RawMessage, error) {
	jsonSlice := json.RawMessage{}
	stringSlice := group.GetClientIdsFromGroup()

	jsonSlice, err := json.Marshal(stringSlice)
	if err != nil {
		return nil, err
	}

	return jsonSlice, nil
}

func GetCurrentGroup(clientId string, s *ChatService) (*Group, *Client, error) {
	client, err := s.GetClient(clientId)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: client (probably) already deleted", err)
	}

	groupId := client.GetGroupId()
	if groupId == "" {
		return nil, client, nil
	}

	group, err := s.GetGroup(groupId)
	if err != nil {
		return nil, client, fmt.Errorf("%w: group not found", err)
	}

	return group, client, nil
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
	if strings.TrimSpace(msg.Content) == "" {
		return nil, fmt.Errorf("%v: missing identifier", ty.ErrNotAvailable)
	}

	identifier := strings.Fields(msg.Content)[0]
	content, _ := strings.CutPrefix(msg.Content, fmt.Sprintf("%s ", identifier))
	msg.Plugin = identifier
	msg.Content = content

	return msg, nil
}
