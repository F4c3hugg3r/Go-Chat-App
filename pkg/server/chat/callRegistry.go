package chat

import (
	"fmt"

	ty "github.com/F4c3hugg3r/Go-Chat-Server/pkg/shared"
)

type CallInterface interface {
	Execute(message *ty.Message) (*ty.Response, error)
}

type CallRegistry struct {
	plugins map[string]CallInterface
}

// RegisterPlugins sets up all the plugins
func RegisterCallPlugins(chatService *ChatService) *CallRegistry {
	cr := &CallRegistry{plugins: make(map[string]CallInterface)}
	cr.plugins[fmt.Sprint("/", ty.OfferSignalFlag)] = NewOfferSignalPlugin(chatService)
	cr.plugins[fmt.Sprint("/", ty.AnswerSignalFlag)] = NewAnswerSignalPlugin(chatService)
	cr.plugins[fmt.Sprint("/", ty.ICECandidateFlag)] = NewICECandidatePlugin(chatService)
	cr.plugins[fmt.Sprint("/", ty.StableSignalFlag)] = NewStableSignalPlugin(chatService)
	cr.plugins[fmt.Sprint("/", ty.ConnectedFlag)] = NewConnectedPlugin(chatService)
	cr.plugins[fmt.Sprint("/", ty.FailedConnectionFlag)] = NewFailedConnectionPlugin(chatService)

	return cr
}

func (pr *CallRegistry) FindAndExecute(message *ty.Message) (*ty.Response, error) {
	plugin, ok := pr.plugins[message.Plugin]
	if !ok {
		return &ty.Response{Err: fmt.Sprintf("%v: no such call plugin found: %s", ty.ErrNoPermission, message.Plugin)}, nil
	}

	return plugin.Execute(message)
}
