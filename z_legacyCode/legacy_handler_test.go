package legacy

// package api

// import (
// 	"bytes"
// 	"encoding/json"
// 	"io"
// 	"net/http"
// 	"net/http/httptest"
// 	"testing"
// 	"time"

// 	"github.com/stretchr/testify/assert"

// 	chat "github.com/F4c3hugg3r/Go-Chat-Server/pkg/server/chat"
// 	ty "github.com/F4c3hugg3r/Go-Chat-Server/pkg/server/types"
// )

// func TestHandleRegistry(t *testing.T) {
// 	service := chat.NewChatService(100)
// 	plugin := chat.RegisterPlugins(service)
// 	handler := NewServerHandler(service, plugin)
// 	service.clients[ty.ClientId] = &chat.Client{
// 		Name:      chat.Name,
// 		ClientId:  chat.ClientId,
// 		ClientCh:  make(chan *ty.Response, 100),
// 		Active:    false,
// 		AuthToken: chat.AuthToken,
// 		LastSign:  time.Now().UTC(),
// 	}

// 	jsonMessage, _ := json.Marshal(chat.DummyExamples[7].Message)
// 	req := httptest.NewRequest(chat.DummyExamples[7].Method, "/users/{clientId}", bytes.NewReader(jsonMessage))
// 	req.SetPathValue("clientId", chat.DummyExamples[7].ClientId)
// 	rec := httptest.NewRecorder()

// 	handler.HandleRegistry(rec, req)
// 	res := rec.Result()
// 	body, _ := io.ReadAll(res.Body)
// 	defer res.Body.Close()

// 	var rsp *ty.Response
// 	dec := json.NewDecoder(bytes.NewReader(body))
// 	dec.Decode(&rsp)

// 	if rsp.Name != "authToken" || rsp.Content == "" {
// 		t.Errorf("response should contain authtoken")
// 	}
// }

// func TestHandleMessages(t *testing.T) {
// 	service := chat.NewChatService(100)
// 	plugin := chat.RegisterPlugins(service)
// 	handler := NewServerHandler(service, plugin)
// 	service.clients[chat.ClientId] = &chat.Client{
// 		Name:      chat.Name,
// 		ClientId:  chat.ClientId,
// 		ClientCh:  make(chan *ty.Response, 100),
// 		Active:    false,
// 		AuthToken: chat.AuthToken,
// 		LastSign:  time.Now().UTC(),
// 	}

// 	for i := 1; i < 6; i++ {
// 		jsonMessage, _ := json.Marshal(chat.DummyExamples[i].Message)
// 		req := httptest.NewRequest(chat.DummyExamples[i].Method, "/users/{clientId}/run", bytes.NewReader(jsonMessage))
// 		req.SetPathValue("clientId", chat.DummyExamples[i].ClientId)
// 		rec := httptest.NewRecorder()

// 		handler.HandleMessages(rec, req)
// 		res := rec.Result()
// 		defer res.Body.Close()

// 		switch i {
// 		case 2:
// 			{
// 				if res.StatusCode != http.StatusOK {
// 					t.Errorf("Status should be %v but was %v instead", http.StatusOK, res.StatusCode)
// 				}
// 			}
// 		case 5:
// 			{
// 				if res.StatusCode != http.StatusInternalServerError {
// 					t.Errorf("Status should be %v but was %v instead", http.StatusInternalServerError, res.StatusCode)
// 				}
// 			}
// 		default:
// 			{
// 				if res.StatusCode != http.StatusBadRequest {
// 					t.Errorf("Status should be %v but was %v instead", http.StatusBadRequest, res.StatusCode)
// 				}
// 			}
// 		}
// 	}
// 	for i := 6; i < 9; i++ {
// 		if i == 7 {
// 			i = 8
// 		}
// 		jsonMessage, _ := json.Marshal(chat.DummyExamples[i].Message)
// 		req := httptest.NewRequest(chat.DummyExamples[i].Method, "/users/{clientId}", bytes.NewReader(jsonMessage))
// 		req.SetPathValue("clientId", chat.DummyExamples[i].ClientId)
// 		rec := httptest.NewRecorder()

// 		handler.HandleMessages(rec, req)
// 		res := rec.Result()
// 		body, _ := io.ReadAll(res.Body)
// 		defer res.Body.Close()

// 		var rsp *ty.Response
// 		dec := json.NewDecoder(bytes.NewReader(body))
// 		dec.Decode(&rsp)

// 		switch i {
// 		case 8:
// 			{
// 				select {
// 				case rsp = <-service.clients[chat.ClientId].ClientCh:
// 					for rsp.Name != "Server" {
// 						rsp = <-service.clients[chat.ClientId].ClientCh
// 					}
// 					if rsp.Name != "Server" {
// 						t.Errorf("there should be a server response")
// 					}
// 				default:
// 					t.Errorf("there should be a response")
// 				}
// 			}
// 		default:
// 			{
// 				if res.StatusCode != http.StatusOK {
// 					t.Errorf("Status should be %v but was %v instead", http.StatusBadRequest, res.StatusCode)
// 				}
// 			}
// 		}
// 	}
// }

// func TestHandleGetRequest(t *testing.T) {
// 	service := chat.NewChatService(100)
// 	plugin := chat.RegisterPlugins(service)
// 	handler := NewServerHandler(service, plugin)

// 	service.clients[ty.ClientId] = &chat.Client{
// 		Name:      chat.Name,
// 		ClientId:  chat.ClientId,
// 		ClientCh:  make(chan *ty.Response, 100),
// 		Active:    false,
// 		AuthToken: chat.AuthToken,
// 		LastSign:  time.Now().UTC(),
// 	}

// 	for i := 0; i < 3; i++ {
// 		service.clients[chat.ClientId].ClientCh <- &chat.DummyResponse
// 		req := httptest.NewRequest(chat.DummyExamples[i].Method, "/users/{clientId}/chat", nil)
// 		req.SetPathValue("clientId", chat.DummyExamples[i].ClientId)
// 		rec := httptest.NewRecorder()

// 		handler.HandleGetRequest(rec, req)
// 		res := rec.Result()
// 		defer res.Body.Close()

// 		switch i {
// 		case 0:
// 			{
// 				data, err := io.ReadAll(res.Body)
// 				if err != nil {
// 					t.Errorf("expected error == nil, got %v instead", err)
// 				}

// 				result, err := json.Marshal(chat.DummyResponse)
// 				if err != nil {
// 					t.Errorf("error extractiong json %v", err)
// 				}
// 				if string(data) != string(result) {
// 					t.Errorf("expected body:'Max: What's poppin' but was %s", string(data))
// 				}
// 			}
// 		default:
// 			{
// 				if res.StatusCode != http.StatusBadRequest {
// 					t.Errorf("Status should be %v but was %v instead", http.StatusBadRequest, res.StatusCode)
// 				}
// 			}
// 		}
// 		service.clients[chat.ClientId].ClientCh <- &chat.DummyResponse
// 		req = httptest.NewRequest("GET", "/users/{clientId}/chat", nil)
// 		req.SetPathValue("clientId", chat.ClientId2)
// 		rec = httptest.NewRecorder()

// 		handler.HandleGetRequest(rec, req)
// 		res = rec.Result()
// 		defer res.Body.Close()

// 		assert.Equal(t, res.StatusCode, http.StatusNotFound)
// 	}
// }

// func TestAuthMiddleware(t *testing.T) {
// 	service := chat.NewChatService(100)
// 	plugin := chat.RegisterPlugins(service)
// 	handler := NewServerHandler(service, plugin)

// 	dummyHandler := func(w http.ResponseWriter, r *http.Request) {
// 		w.Write([]byte("success"))
// 	}

// 	service.clients[chat.ClientId] = &chat.Client{
// 		Name:      chat.Name,
// 		ClientId:  chat.ClientId,
// 		ClientCh:  make(chan *ty.Response, 100),
// 		Active:    false,
// 		AuthToken: chat.AuthToken,
// 		LastSign:  time.Now().UTC(),
// 	}

// 	for i := 2; i < 9; i++ {
// 		if i == 3 {
// 			i = 6
// 		}
// 		jsonMessage, _ := json.Marshal(chat.DummyExamples[i].Message)
// 		req := httptest.NewRequest(chat.DummyExamples[i].Method, "/users/{clientId}", bytes.NewReader(jsonMessage))
// 		req.SetPathValue("clientId", chat.DummyExamples[i].ClientId)
// 		req.Header.Set("Authorization", chat.DummyExamples[i].Token)
// 		rec := httptest.NewRecorder()

// 		authHandler := handler.AuthMiddleware(dummyHandler)
// 		authHandler(rec, req)
// 		res := rec.Result()
// 		defer res.Body.Close()

// 		switch i {
// 		case 2:
// 			{
// 				data, err := io.ReadAll(res.Body)
// 				if err != nil {
// 					t.Errorf("expected error == nil, got %v instead", err)
// 				}

// 				if string(data) != "success" {
// 					t.Errorf("expected body:'success' but was %s", string(data))
// 				}
// 			}
// 		case 6:
// 			{
// 				if res.StatusCode != http.StatusBadRequest {
// 					t.Errorf("Status should be %v but was %v instead", http.StatusBadRequest, res.StatusCode)
// 				}
// 			}
// 		default:
// 			{
// 				if res.StatusCode != http.StatusForbidden {
// 					t.Errorf("Status should be %v but was %v instead", http.StatusForbidden, res.StatusCode)
// 				}
// 			}
// 		}
// 	}
// }
