package server

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHelpPlugin(t *testing.T) {
	service := NewChatService(100)
	registry := RegisterPlugins(service)
	message := &Message{Name: "Arndt", Content: "wubbalubbadubdub", Plugin: "/help", ClientId: dummyClient.ClientId}

	rsp, err := registry.FindAndExecute(message)
	assert.Nil(t, err)
	assert.Equal(t, rsp, Response{Name: "Help", Content: "[{\"Command\":\"/help\",\"Description\":\"tells every plugin " +
		"and their description\"},{\"Command\":\"/time\",\"Description\":\"tells you the current time\"},{\"Command\":\"/users" +
		"\",\"Description\":\"tells you information about all the current users\"},{\"Command\":\"/private\",\"Description\":" +
		"\"lets you send a private message to someone - template: '/private {Id} {message}'\"},{\"Command\":\"/quit\"," +
		"\"Description\":\"loggs you out of the chat\"}]"})

}

func TestUserPlugin(t *testing.T) {
	service := NewChatService(100)
	registry := RegisterPlugins(service)
	message := &Message{Name: "Arndt", Content: "wubbalubbadubdub", Plugin: "/users", ClientId: dummyClient.ClientId}

	rsp, err := registry.FindAndExecute(message)
	assert.Equal(t, rsp, Response{Name: "Users", Content: "[]"})

	service.clients[clientId] = dummyClient
	service.clients[clientId2] = dummyClient2

	rsp, err = registry.FindAndExecute(message)
	assert.Nil(t, err)
	assert.Equal(t, rsp, Response{Name: "Users", Content: "[{\"Name\":\"Max\",\"ClientId\"" +
		":\"clientId-DyGWNnLrLWnbuhf-LgBUAdAxdZf-U1pgRw\",\"Active\":false},{\"Name\":\"Len\"" +
		",\"ClientId\":\"clientId2-yGWNnLrLWnbuhf-LgBUAdAxdZf-U1pgRw\",\"Active\":true}]"})
}

func TestRegisterPlugin(t *testing.T) {
	service := NewChatService(100)
	registry := RegisterPlugins(service)
	message := &Message{Name: "Arndt", Content: "wubbalubbadubdub", Plugin: "/register", ClientId: dummyClient.ClientId}

	service.clients[clientId] = dummyClient

	_, err := registry.FindAndExecute(message)
	assert.Error(t, err, NoPermissionError)

	delete(service.clients, clientId)

	rsp, err := registry.FindAndExecute(message)
	assert.Nil(t, err)
	assert.Equal(t, rsp.Name, "authToken")
	assert.NotEmpty(t, rsp.Content)
}

func TestQuitPlugin(t *testing.T) {
	message := &Message{Name: "Arndt", Content: "wubbalubbadubdub", Plugin: "/quit", ClientId: dummyClient.ClientId}
	service := NewChatService(100)
	registry := RegisterPlugins(service)

	_, err := registry.FindAndExecute(message)
	assert.Error(t, err, ClientNotAvailableError)

	service.clients[clientId] = dummyClient

	rsp, err := registry.FindAndExecute(message)
	assert.Equal(t, 0, len(service.clients))
	assert.Nil(t, err)
	assert.Equal(t, rsp, Response{Name: message.Name, Content: "logged out"})
}

func TestBroadcastPlugin(t *testing.T) {
	service := NewChatService(100)
	registry := RegisterPlugins(service)
	message := &Message{Name: "Arndt", Content: "wubbalubbadubdub", Plugin: "/broadcast"}

	_, err := registry.FindAndExecute(message)
	assert.ErrorIs(t, err, ClientNotAvailableError)

	service.clients[clientId] = dummyClient
	service.clients[clientId2] = dummyClient2

	go registry.FindAndExecute(message)

	for _, client := range service.clients {
		go func() {
			select {
			case <-client.clientCh:
			case <-time.After(3 * time.Second):
				t.Error("message wasn't broadcastet")
			}
		}()
	}

	_, err = registry.FindAndExecute(&Message{Name: "Arndt", Content: "", Plugin: "/broadcast"})
	assert.ErrorIs(t, err, EmptyStringError)
}
