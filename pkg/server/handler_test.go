package server

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

var (
	name          = "Max"
	clientId      = "clientId-DyGWNnLrLWnbuhf-LgBUAdAxdZf-U1pgRw"
	clientId2     = "clientId2-yGWNnLrLWnbuhf-LgBUAdAxdZf-U1pgRw"
	authToken     = "authId-5EDyGWNnLrLWnbuhf-LgBUAdAxdZf-U1pgRwc7ex1dt5EDyGWNnLrLWnbuhf-LgBUAdAxdZf-U1pgRw"
	authToken2    = "authId2-EDyGWNnLrLWnbuhf-LgBUAdAxdZf-U1pgRwc7ex1dt5EDyGWNnLrLWnbuhf-LgBUAdAxdZf-U1pgRw"
	dummyClient   = &Client{name, clientId, make(chan Response), true, authToken, time.Now()}
	dummyMessage  = Message{Name: name, Content: "What's poppin"}
	dummyResponse = Response{Name: name, Content: "What's poppin"}
	dummyExamples = []dummyRequests{
		{
			method:   "GET",
			clientId: clientId,
		},
		{
			method:   "GET",
			clientId: "",
		},
		{
			method:   "POST",
			clientId: clientId,
			body:     name,
			token:    authToken,
		},
		{
			method:   "POST",
			clientId: "",
			body:     name,
		},
		{
			method:   "POST",
			clientId: clientId,
			body:     "",
		},
		{
			method:   "POST",
			clientId: clientId2,
			body:     strings.Repeat("s", (int(1<<20) + 1)),
		},
		{
			method:   "POST",
			clientId: clientId,
			token:    "",
		},
		{
			method:   "POST",
			clientId: clientId2,
			token:    authToken,
		},
		{
			method:   "POST",
			clientId: clientId,
			token:    authToken2,
		},
	}
)

func registerClient(clientId string, body Message) (token string, e error) { return authToken, nil }
func (handler *ServerHandler) dummyHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("success"))
}

type dummyRequests struct {
	method   string
	clientId string
	body     string
	token    string
}

func TestAuthMiddleware(t *testing.T) {
	service := NewChatService()
	plugin := RegisterPlugins(service)
	handler := NewServerHandler(service, plugin)
	service.clients[clientId] = dummyClient
	for i := 2; i < 9; i++ {
		if i == 3 {
			i = 6
		}
		req := httptest.NewRequest(dummyExamples[i].method, "/users/{clientId}", strings.NewReader(dummyExamples[i].body))
		req.SetPathValue("clientId", dummyExamples[i].clientId)
		req.Header.Set("Authorization", dummyExamples[i].token)
		rec := httptest.NewRecorder()

		authHandler := handler.AuthMiddleware(handler.dummyHandler)
		authHandler(rec, req)
		res := rec.Result()
		defer res.Body.Close()

		switch i {
		case 2:
			{
				data, err := io.ReadAll(res.Body)
				if err != nil {
					t.Errorf("expected error == nil, got %v instead", err)
				}

				if string(data) != "success" {
					t.Errorf("expected body:'success' but was %s", string(data))
				}
			}
		case 6:
			{
				if res.StatusCode != http.StatusBadRequest {
					t.Errorf("Status should be %v but was %v instead", http.StatusBadRequest, res.StatusCode)
				}
			}
		default:
			{
				if res.StatusCode != http.StatusForbidden {
					t.Errorf("Status should be %v but was %v instead", http.StatusForbidden, res.StatusCode)
				}
			}
		}
	}

}

// läuft noch nicht weil Message ohne Plugin übergeben wird
func TestHandleMessages(t *testing.T) {
	service := NewChatService()
	plugin := RegisterPlugins(service)
	handler := NewServerHandler(service, plugin)
	broadcastTest := 0
	handler.broadcaster = func(msg *Message) { broadcastTest += 1 }
	for i := 1; i < 6; i++ {
		req := httptest.NewRequest(dummyExamples[i].method, "/users/{clientId}/message", strings.NewReader(dummyExamples[i].body))
		req.SetPathValue("clientId", dummyExamples[i].clientId)
		rec := httptest.NewRecorder()

		handler.HandleRegistry(rec, req)
		res := rec.Result()
		defer res.Body.Close()

		switch i {
		case 2:
			{
				if res.StatusCode != http.StatusOK && broadcastTest != 1 {
					t.Errorf("Status should be %v but was %v instead. And broadcastTest variable "+
						"should be 1 but was %d", http.StatusOK, res.StatusCode, broadcastTest)
				}
			}
		default:
			{
				if res.StatusCode != http.StatusBadRequest {
					t.Errorf("Status should be %v but was %v instead", http.StatusBadRequest, res.StatusCode)
				}
			}
		}
	}
}

// läuft noch nicht weil Message ohne Plugin übergeben wird
func TestHandleRegistry(t *testing.T) {
	service := NewChatService()
	plugin := RegisterPlugins(service)
	handler := NewServerHandler(service, plugin)
	handler.registerer = registerClient

	for i := 1; i < 6; i++ {
		req := httptest.NewRequest(dummyExamples[i].method, "/users/{clientId}", strings.NewReader(dummyExamples[i].body))
		req.SetPathValue("clientId", dummyExamples[i].clientId)
		rec := httptest.NewRecorder()

		handler.HandleRegistry(rec, req)
		res := rec.Result()
		defer res.Body.Close()

		switch i {
		case 2:
			{
				data, err := io.ReadAll(res.Body)
				if err != nil {
					t.Errorf("expected error == nil, got %v instead", err)
				}

				if string(data) != authToken {
					t.Errorf("expected body: %s but was %s", authToken, string(data))
				}
			}
		default:
			{
				if res.StatusCode != http.StatusBadRequest {
					t.Errorf("Status should be %v but was %v instead", http.StatusBadRequest, res.StatusCode)
				}
			}
		}
	}
}

func TestHandleGetRequest(t *testing.T) {
	service := NewChatService()
	plugin := RegisterPlugins(service)
	handler := NewServerHandler(service, plugin)

	service.clients[clientId] = dummyClient
	go func() {
		time.Sleep(1 * time.Second)
		dummyClient.clientCh <- dummyResponse
	}()

	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(dummyExamples[i].method, "/users/{clientId}/chat", nil)
		req.SetPathValue("clientId", dummyExamples[i].clientId)
		rec := httptest.NewRecorder()

		handler.HandleGetRequest(rec, req)
		res := rec.Result()
		defer res.Body.Close()

		switch i {
		case 0:
			{
				data, err := io.ReadAll(res.Body)
				if err != nil {
					t.Errorf("expected error == nil, got %v instead", err)
				}

				comparator, err := json.Marshal(dummyResponse)
				if err != nil {
					t.Errorf("error extractiong json %v", err)
				}
				if string(data) != string(comparator) {
					t.Errorf("expected body:'Max: What's poppin' but was %s", string(data))
				}
			}
		default:
			{
				if res.StatusCode != http.StatusBadRequest {
					t.Errorf("Status should be %v but was %v instead", http.StatusBadRequest, res.StatusCode)
				}
			}
		}
	}
}
