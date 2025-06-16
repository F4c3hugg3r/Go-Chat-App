package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

var (
	service       = NewChatService()
	handler       = NewServerHandler(service)
	clientId      = "clientId-DyGWNnLrLWnbuhf-LgBUAdAxdZf-U1pgRw"
	clientId2     = "clientId2-yGWNnLrLWnbuhf-LgBUAdAxdZf-U1pgRw"
	authToken     = "authId-5EDyGWNnLrLWnbuhf-LgBUAdAxdZf-U1pgRwc7ex1dt5EDyGWNnLrLWnbuhf-LgBUAdAxdZf-U1pgRw"
	authToken2    = "authId2-EDyGWNnLrLWnbuhf-LgBUAdAxdZf-U1pgRwc7ex1dt5EDyGWNnLrLWnbuhf-LgBUAdAxdZf-U1pgRw"
	dummyClient   = &Client{"Max", clientId, make(chan Message), true, authToken}
	dummyMessage  = Message{name: "Max", content: "What's poppin"}
	dummyExamples = []dummyRequests{
		{
			method:    "GET",
			clientId:  clientId,
			authToken: authToken,
		},
		{
			method:    "GET",
			clientId:  "",
			authToken: authToken,
		},
		{
			method:    "POST",
			clientId:  clientId,
			authToken: authToken,
		},
	}
)

type dummyRequests struct {
	method    string
	clientId  string
	authToken string
}

func TestHandleRegistry(t *testing.T) {
	//TODO
}

func TestHandleGetRequest(t *testing.T) {
	service.clients[clientId] = dummyClient
	go func() {
		time.Sleep(1 * time.Second)
		dummyClient.clientCh <- dummyMessage
	}()

	for i := range dummyExamples {
		req := httptest.NewRequest(dummyExamples[i].method, "/users/{clientId}/chat", nil)
		req.SetPathValue("clientId", dummyExamples[i].clientId)
		rec := httptest.NewRecorder()

		handler.HandleGetRequest(rec, req)
		res := rec.Result()
		defer res.Body.Close()

		if i == 0 {
			data, err := io.ReadAll(res.Body)
			if err != nil {
				t.Errorf("expected error == nil, got %v instead", err)
			}

			if string(data) != "Max: What's poppin" {
				t.Errorf("expected body:'Max: What's poppin' but was %s", string(data))
			}
		}

		if i == 1 {
			if res.StatusCode != http.StatusBadRequest {
				t.Errorf("Status should be %v but was %v instead", http.StatusBadRequest, res.StatusCode)
			}
		}

		if i == 2 {
			if res.StatusCode != http.StatusBadRequest {
				t.Errorf("Status should be %v but was %v instead", http.StatusBadRequest, res.StatusCode)
			}
		}
	}
}
