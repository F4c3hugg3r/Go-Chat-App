package server

import (
	"testing"
	"time"
)

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
