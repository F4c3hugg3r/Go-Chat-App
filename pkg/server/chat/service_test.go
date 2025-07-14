package chat

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	ty "github.com/F4c3hugg3r/Go-Chat-Server/pkg/server/types"
)

func TestDecodeToMessage(t *testing.T) {
	fakeBody := []byte("fake")
	_, err := DecodeToMessage(fakeBody)
	assert.Error(t, err)

	message := ty.Message{Name: "Arndt", Content: "wubbalubbadubdub", Plugin: "/broadcast"}

	jsonMessage, err := json.Marshal(&message)
	if err != nil {
		t.Errorf("%v: Message couldn't be parsed to json", err)
	}

	resultMessage, err := DecodeToMessage(jsonMessage)
	assert.Nil(t, err)
	assert.Equal(t, resultMessage, message)
}

func TestEcho(t *testing.T) {
	service := NewChatService(100)

	err := service.Echo(ClientId, &DummyResponse)
	assert.ErrorIs(t, err, ty.ErrClientNotAvailable)

	service.clients[ClientId] = &Client{
		Name:      Name,
		ClientId:  ClientId,
		clientCh:  make(chan *ty.Response, 100),
		Active:    false,
		authToken: AuthToken,
		lastSign:  time.Now().UTC().Add(-time.Hour),
	}
	err = service.Echo(ClientId, &DummyResponse)
	assert.Nil(t, err)

	select {
	case <-service.clients[ClientId].clientCh:
	default:
		t.Errorf("client should receive a message")
	}

}

func TestInactiveClientDeleter(t *testing.T) {
	service := NewChatService(100)
	service.clients[ClientId] = &Client{
		Name:      Name,
		ClientId:  ClientId,
		clientCh:  make(chan *ty.Response, 100),
		Active:    false,
		authToken: AuthToken,
		lastSign:  time.Now().UTC().Add(-time.Hour),
	}
	if len(service.clients) != 1 {
		t.Errorf(("Setup incorrect there should be just 1 client but there is %d"), len(service.clients))
	}

	service.InactiveClientDeleter(30 * time.Minute)
	if len(service.clients) != 0 {
		t.Errorf(("There should be 0 client but there is %d"), len(service.clients))
	}
}

func TestGetClient(t *testing.T) {
	service := NewChatService(100)
	_, err := service.GetClient(ClientId)
	if err == nil {
		t.Error("there should be an error but instead it's nil")
	}

	dummyClient := &Client{
		Name:      Name,
		ClientId:  ClientId,
		clientCh:  make(chan *ty.Response, 100),
		Active:    false,
		authToken: AuthToken,
		lastSign:  time.Now().UTC().Add(-time.Hour),
	}

	service.clients[ClientId] = dummyClient
	client, err := service.GetClient(ClientId)
	if err != nil {
		t.Errorf("error should be nil but instead is %v", err)
	}
	if client != dummyClient {
		t.Errorf("client should be %v but instead was %v", dummyClient, client)
	}
}
