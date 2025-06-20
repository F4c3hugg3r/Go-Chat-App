package server

import (
	"testing"
	"time"
)

var (
	name2               = "Len"
	dummyClientInactive = &Client{
		name:      name,
		clientId:  clientId,
		clientCh:  make(chan Message),
		active:    false,
		authToken: authToken,
	}
	dummyClient2 = &Client{
		name:      name2,
		clientId:  clientId2,
		clientCh:  make(chan Message),
		active:    true,
		authToken: authToken2,
	}
)

func TestLogOutClient(t *testing.T) {
	service.clients[clientId] = dummyClient
	if len(service.clients) != 1 {
		t.Errorf(("Setup incorrect there should be just 1 client but there is %d"), len(service.clients))
	}

	err := service.logOutClient(clientId)
	if len(service.clients) != 0 || err != nil {
		t.Errorf(("There should be 0 client but there is %d"), len(service.clients))
	}

	err = service.logOutClient(clientId)
	if len(service.clients) != 0 || err == nil {
		t.Errorf(("There should be 0 client but there is %d"), len(service.clients))
	}
}

func TestInactiveClientDeleter(t *testing.T) {
	service.clients[clientId] = dummyClientInactive
	if len(service.clients) != 1 {
		t.Errorf(("Setup incorrect there should be just 1 client but there is %d"), len(service.clients))
	}

	service.InactiveClientDeleter()
	if len(service.clients) != 0 {
		t.Errorf(("There should be 0 client but there is %d"), len(service.clients))
	}
}

func TestRegisterClient(t *testing.T) {
	service.clients[clientId] = dummyClientInactive
	if len(service.clients) != 1 {
		t.Errorf(("Setup incorrect there should be just 1 client but there is %d"), len(service.clients))
	}
	token, err := service.registerClient(clientId2, name)

	if len(service.clients) != 2 {
		t.Errorf(("There should be just 1 client but there is %d"), len(service.clients))
	}

	if err != nil {
		t.Errorf("err should be nil but is %v", err)
	}
	if token == "" || len(token) != 86 {
		t.Errorf("token should be 86 chars long but is %d: %s", len(token), token)
	}

	_, err = service.registerClient(clientId, name)
	if err == nil {
		t.Error("there should be an error but instead is nil")
	}
}

func TestSendBroadcast(t *testing.T) {
	service.clients[clientId] = dummyClient
	service.clients[clientId2] = dummyClient2

	go func() {
		time.Sleep(time.Second)
		service.sendBroadcast(Message{name: "Arndt", content: "wubbalubbadubdub"})
	}()

	select {
	case <-service.clients[clientId].clientCh:
		select {
		case <-service.clients[clientId2].clientCh:

		case <-time.After(15 * time.Second):
			t.Error("message wasn't broadcastet")
		}
	case <-service.clients[clientId2].clientCh:
		select {
		case <-service.clients[clientId].clientCh:

		case <-time.After(15 * time.Second):
			t.Error("message wasn't broadcastet")
		}
	case <-time.After(15 * time.Second):
		t.Error("message wasn't broadcastet")
	}
}

func TestGetClient(t *testing.T) {
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
