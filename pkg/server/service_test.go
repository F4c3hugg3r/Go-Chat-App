package server

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDecodeToMessage(t *testing.T) {
	fakeBody := []byte("fake")
	_, err := DecodeToMessage(fakeBody)
	assert.Error(t, err)

	message := Message{Name: "Arndt", Content: "wubbalubbadubdub", Plugin: "/broadcast"}

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

	err := service.echo(clientId, dummyResponse)
	assert.ErrorIs(t, err, ClientNotAvailableError)

	service.clients[clientId] = dummyClientInactive
	err = service.echo(clientId, dummyResponse)
	assert.Nil(t, err)

	select {
	case <-service.clients[clientId].clientCh:
		assert.True(t, service.clients[clientId].Active)
	case <-time.After(500 * time.Millisecond):
		t.Errorf("client should receive a message")
	}

}

func TestInactiveClientDeleter(t *testing.T) {
	service := NewChatService(100)
	service.clients[clientId] = dummyClientInactive
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
	_, err := service.getClient(clientId)
	if err == nil {
		t.Error("there should be an error but instead it's nil")
	}

	service.clients[clientId] = dummyClient
	client, err := service.getClient(clientId)
	if err != nil {
		t.Errorf("error should be nil but instead is %v", err)
	}
	if client != dummyClient {
		t.Errorf("client should be %v but instead was %v", dummyClient, client)
	}
}
