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
	assert.Equal(t, rsp.Name, "Help")
}

func TestUserPlugin(t *testing.T) {
	service := NewChatService(100)
	registry := RegisterPlugins(service)
	message := &Message{Name: "Arndt", Content: "wubbalubbadubdub", Plugin: "/users", ClientId: dummyClient.ClientId}

	rsp, err := registry.FindAndExecute(message)
	assert.Equal(t, rsp, &Response{Name: "Users", Content: "[]"})

	service.clients[clientId] = &Client{
		Name:      name,
		ClientId:  clientId,
		clientCh:  make(chan *Response, 100),
		Active:    false,
		authToken: authToken,
		lastSign:  time.Now().UTC(),
	}
	service.clients[clientId2] = &Client{
		Name:      name2,
		ClientId:  clientId2,
		clientCh:  make(chan *Response, 100),
		Active:    true,
		authToken: authToken2,
		lastSign:  time.Now().UTC(),
	}

	rsp, err = registry.FindAndExecute(message)
	assert.Nil(t, err)
	assert.Equal(t, rsp, &Response{Name: "Users", Content: "[{\"Name\":\"Arndt\",\"ClientId\"" +
		":\"clientId-DyGWNnLrLWnbuhf-LgBUAdAxdZf-U1pgRw\",\"Active\":false},{\"Name\":\"Len\"" +
		",\"ClientId\":\"clientId2-yGWNnLrLWnbuhf-LgBUAdAxdZf-U1pgRw\",\"Active\":true}]"})
}

func TestRegisterPlugin(t *testing.T) {
	service := NewChatService(100)
	registry := RegisterPlugins(service)
	message := &Message{Name: "Arndt", Content: "wubbalubbadubdub", Plugin: "/register", ClientId: dummyClient.ClientId}

	service.clients[clientId] = &Client{
		Name:      name,
		ClientId:  clientId,
		clientCh:  make(chan *Response, 100),
		Active:    false,
		authToken: authToken,
		lastSign:  time.Now().UTC(),
	}

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

	service.clients[clientId] = &Client{
		Name:      name,
		ClientId:  clientId,
		clientCh:  make(chan *Response, 100),
		Active:    false,
		authToken: authToken,
		lastSign:  time.Now().UTC(),
	}

	rsp, err := registry.FindAndExecute(message)
	assert.Equal(t, 0, len(service.clients))
	assert.Nil(t, err)
	assert.Equal(t, rsp, &Response{Name: message.Name, Content: "logged out"})
}

func TestBroadcastPlugin(t *testing.T) {
	service := NewChatService(100)
	registry := RegisterPlugins(service)
	message := &Message{Name: "Arndt", Content: "wubbalubbadubdub", Plugin: "/broadcast"}

	_, err := registry.FindAndExecute(message)
	assert.ErrorIs(t, err, ClientNotAvailableError)

	service.clients[clientId] = &Client{
		Name:      name,
		ClientId:  clientId,
		clientCh:  make(chan *Response, 100),
		Active:    false,
		authToken: authToken,
		lastSign:  time.Now().UTC(),
	}
	service.clients[clientId2] = &Client{
		Name:      name2,
		ClientId:  clientId2,
		clientCh:  make(chan *Response, 100),
		Active:    true,
		authToken: authToken2,
		lastSign:  time.Now().UTC(),
	}

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
