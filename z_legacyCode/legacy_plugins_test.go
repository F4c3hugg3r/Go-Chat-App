package legacy

// package chat

// import (
// 	"testing"
// 	"time"

// 	ty "github.com/F4c3hugg3r/Go-Chat-Server/pkg/shared"

// 	"github.com/stretchr/testify/assert"
// )

// func TestHelpPlugin(t *testing.T) {
// 	service := NewChatService(100)
// 	registry := RegisterPlugins(service)
// 	message := &ty.Message{Name: "Arndt", Content: "wubbalubbadubdub", Plugin: "/help", ClientId: DummyClient.ClientId}

// 	rsp, err := registry.FindAndExecute(message)
// 	assert.Nil(t, err)
// 	assert.Equal(t, rsp.RspName, "Help")
// }

// func TestUserPlugin(t *testing.T) {
// 	service := NewChatService(100)
// 	registry := RegisterPlugins(service)
// 	message := &ty.Message{Name: "Arndt", Content: "wubbalubbadubdub", Plugin: "/users", ClientId: DummyClient.ClientId}

// 	rsp, err := registry.FindAndExecute(message)
// 	assert.Equal(t, rsp, &ty.Response{RspName: "Users", Content: "[]"})

// 	service.clients[ClientId] = &Client{
// 		Name:      ClientName,
// 		ClientId:  ClientId,
// 		clientCh:  make(chan *ty.Response, 100),
// 		Active:    false,
// 		authToken: AuthToken,
// 		lastSign:  time.Now().UTC(),
// 	}
// 	service.clients[ClientId2] = &Client{
// 		Name:      ClientName2,
// 		ClientId:  ClientId2,
// 		clientCh:  make(chan *ty.Response, 100),
// 		Active:    true,
// 		authToken: AuthToken2,
// 		lastSign:  time.Now().UTC(),
// 	}

// 	rsp, err = registry.FindAndExecute(message)
// 	assert.Nil(t, err)
// 	assert.Equal(t, rsp.RspName, "Users")
// 	assert.Contains(t, rsp.Content, "{\"Name\":\"Len\""+
// 		",\"ClientId\":\"clientId2-yGWNnLrLWnbuhf-LgBUAdAxdZf-U1pgRw\",\"Active\":true}")
// 	assert.Contains(t, rsp.Content, "{\"Name\":\"Arndt\",\"ClientId\""+
// 		":\"clientId-DyGWNnLrLWnbuhf-LgBUAdAxdZf-U1pgRw\",\"Active\":false}")
// }

// func TestRegisterPlugin(t *testing.T) {
// 	service := NewChatService(100)
// 	registry := RegisterPlugins(service)
// 	message := &ty.Message{Name: "Arndt", Content: "wubbalubbadubdub", Plugin: "/register", ClientId: DummyClient.ClientId}

// 	service.clients[ClientId] = &Client{
// 		Name:      ClientName,
// 		ClientId:  ClientId,
// 		clientCh:  make(chan *ty.Response, 100),
// 		Active:    false,
// 		authToken: AuthToken,
// 		lastSign:  time.Now().UTC(),
// 	}

// 	_, err := registry.FindAndExecute(message)
// 	assert.Error(t, err, ty.ErrNoPermission)

// 	delete(service.clients, ClientId)

// 	rsp, err := registry.FindAndExecute(message)
// 	assert.Nil(t, err)
// 	assert.Equal(t, rsp.RspName, "authToken")
// 	assert.NotEmpty(t, rsp.Content)
// }

// func TestQuitPlugin(t *testing.T) {
// 	message := &ty.Message{Name: "Arndt", Content: "wubbalubbadubdub", Plugin: "/quit", ClientId: DummyClient.ClientId}
// 	service := NewChatService(100)
// 	registry := RegisterPlugins(service)

// 	_, err := registry.FindAndExecute(message)
// 	assert.Error(t, err, ty.ErrNotAvailable)

// 	service.clients[ClientId] = &Client{
// 		Name:      ClientName,
// 		ClientId:  ClientId,
// 		clientCh:  make(chan *ty.Response, 100),
// 		Active:    false,
// 		authToken: AuthToken,
// 		lastSign:  time.Now().UTC(),
// 	}

// 	rsp, err := registry.FindAndExecute(message)
// 	assert.Equal(t, 0, len(service.clients))
// 	assert.Nil(t, err)
// 	assert.Equal(t, rsp, &ty.Response{RspName: message.Name, Content: "logged out"})
// }

// func TestBroadcastPlugin(t *testing.T) {
// 	service := NewChatService(100)
// 	registry := RegisterPlugins(service)
// 	message := &ty.Message{Name: "Arndt", Content: "wubbalubbadubdub", Plugin: "/broadcast"}

// 	_, err := registry.FindAndExecute(message)
// 	assert.ErrorIs(t, err, ty.ErrNotAvailable)

// 	service.clients[ClientId] = &Client{
// 		Name:      ClientName,
// 		ClientId:  ClientId,
// 		clientCh:  make(chan *ty.Response, 100),
// 		Active:    false,
// 		authToken: AuthToken,
// 		lastSign:  time.Now().UTC(),
// 	}
// 	service.clients[ClientId2] = &Client{
// 		Name:      ClientName2,
// 		ClientId:  ClientId2,
// 		clientCh:  make(chan *ty.Response, 100),
// 		Active:    true,
// 		authToken: AuthToken2,
// 		lastSign:  time.Now().UTC(),
// 	}

// 	go registry.FindAndExecute(message)

// 	for _, client := range service.clients {
// 		go func() {
// 			select {
// 			case <-client.clientCh:
// 			case <-time.After(3 * time.Second):
// 				t.Error("message wasn't broadcastet")
// 			}
// 		}()
// 	}

// 	_, err = registry.FindAndExecute(&ty.Message{Name: "Arndt", Content: "", Plugin: "/broadcast"})
// 	assert.ErrorIs(t, err, ty.ErrEmptyString)
// }
