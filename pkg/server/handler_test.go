package server

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHandleMessages(t *testing.T) {
	service := NewChatService(100)
	plugin := RegisterPlugins(service)
	handler := NewServerHandler(service, plugin)
	service.clients[clientId] = &Client{
		Name:      name,
		ClientId:  clientId,
		clientCh:  make(chan *Response, 100),
		Active:    false,
		authToken: authToken,
		lastSign:  time.Now(),
	}

	for i := 1; i < 6; i++ {
		jsonMessage, _ := json.Marshal(dummyExamples[i].message)
		req := httptest.NewRequest(dummyExamples[i].method, "/users/{clientId}/run", bytes.NewReader(jsonMessage))
		req.SetPathValue("clientId", dummyExamples[i].clientId)
		rec := httptest.NewRecorder()

		handler.HandleMessages(rec, req)
		res := rec.Result()
		defer res.Body.Close()

		switch i {
		case 2:
			{
				if res.StatusCode != http.StatusOK {
					t.Errorf("Status should be %v but was %v instead", http.StatusOK, res.StatusCode)
				}
			}
		case 5:
			{
				if res.StatusCode != http.StatusInternalServerError {
					t.Errorf("Status should be %v but was %v instead", http.StatusInternalServerError, res.StatusCode)
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
	for i := 7; i < 9; i++ {
		jsonMessage, _ := json.Marshal(dummyExamples[i].message)
		req := httptest.NewRequest(dummyExamples[i].method, "/users/{clientId}", bytes.NewReader(jsonMessage))
		req.SetPathValue("clientId", dummyExamples[i].clientId)
		rec := httptest.NewRecorder()

		handler.HandleMessages(rec, req)
		res := rec.Result()
		body, _ := io.ReadAll(res.Body)
		defer res.Body.Close()

		var rsp *Response
		dec := json.NewDecoder(bytes.NewReader(body))
		dec.Decode(&rsp)

		switch i {
		case 7:
			{
				if rsp.Name != "authToken" || rsp.Content == "" {
					t.Errorf("response should contain authtoken")
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
	service := NewChatService(100)
	plugin := RegisterPlugins(service)
	handler := NewServerHandler(service, plugin)

	service.clients[clientId] = &Client{
		Name:      name,
		ClientId:  clientId,
		clientCh:  make(chan *Response, 100),
		Active:    false,
		authToken: authToken,
		lastSign:  time.Now(),
	}

	for i := 0; i < 3; i++ {
		service.clients[clientId].clientCh <- &dummyResponse
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

				result, err := json.Marshal(dummyResponse)
				if err != nil {
					t.Errorf("error extractiong json %v", err)
				}
				if string(data) != string(result) {
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
		service.clients[clientId].clientCh <- &dummyResponse
		req = httptest.NewRequest("GET", "/users/{clientId}/chat", nil)
		req.SetPathValue("clientId", clientId2)
		rec = httptest.NewRecorder()

		handler.HandleGetRequest(rec, req)
		res = rec.Result()
		defer res.Body.Close()

		assert.Equal(t, res.StatusCode, http.StatusNotFound)
	}
}

func TestAuthMiddleware(t *testing.T) {
	service := NewChatService(100)
	plugin := RegisterPlugins(service)
	handler := NewServerHandler(service, plugin)

	dummyHandler := func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("success"))
	}

	service.clients[clientId] = &Client{
		Name:      name,
		ClientId:  clientId,
		clientCh:  make(chan *Response, 100),
		Active:    false,
		authToken: authToken,
		lastSign:  time.Now(),
	}

	for i := 2; i < 9; i++ {
		if i == 3 {
			i = 6
		}
		jsonMessage, _ := json.Marshal(dummyExamples[i].message)
		req := httptest.NewRequest(dummyExamples[i].method, "/users/{clientId}", bytes.NewReader(jsonMessage))
		req.SetPathValue("clientId", dummyExamples[i].clientId)
		req.Header.Set("Authorization", dummyExamples[i].token)
		rec := httptest.NewRecorder()

		authHandler := handler.AuthMiddleware(dummyHandler)
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
