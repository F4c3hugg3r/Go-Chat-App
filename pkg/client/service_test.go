package client

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	name     = "Arndt"
	clientId = "clientId-DyGWNnLrLWnbuhf-LgBUAdAxdZf-U1pgRw"
	//clientId2  = "clientId2-yGWNnLrLWnbuhf-LgBUAdAxdZf-U1pgRw"
	authToken = "authId-5EDyGWNnLrLWnbuhf-LgBUAdAxdZf-U1pgRwc7ex1dt5EDyGWNnLrLWnbuhf-LgBUAdAxdZf-U1pgRw"
	//authToken2 = "authId2-EDyGWNnLrLWnbuhf-LgBUAdAxdZf-U1pgRwc7ex1dt5EDyGWNnLrLWnbuhf-LgBUAdAxdZf-U1pgRw"
)

func TestRegister(t *testing.T) {
	clientService := &Client{
		clientId:   clientId,
		authToken:  authToken,
		reader:     bufio.NewReader(os.Stdin),
		writer:     io.Writer(os.Stdout),
		HttpClient: &http.Client{},
	}

	rsp := Response{Name: "authToken", Content: authToken}
	jsonRsp, _ := json.Marshal(rsp)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(jsonRsp)
	}))
	defer ts.Close()
	clientService.authToken = ""

	err := clientService.Register(ts.URL)
	assert.Error(t, err)

	ts.URL = fmt.Sprintf("%s/users/%s", ts.URL, clientId)

	clientService.reader = bufio.NewReader(strings.NewReader(name + "\n"))
	err = clientService.Register(ts.URL)
	assert.NoError(t, err)
	assert.Equal(t, clientService.authToken, authToken)
}

func TestGetMessages(t *testing.T) {
	clientService := &Client{
		clientId:   clientId,
		authToken:  authToken,
		reader:     bufio.NewReader(os.Stdin),
		writer:     io.Writer(os.Stdout),
		HttpClient: &http.Client{},
	}

	_, cancel := context.WithCancel(context.Background())

	rsp := Response{Name: "Max", Content: "wubbalubbadubdub"}
	jsonRsp, _ := json.Marshal(rsp)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(time.Second)
		w.Write(jsonRsp)
	}))
	defer ts.Close()

	var output bytes.Buffer
	clientService.writer = &output

	ts.URL = fmt.Sprintf("%s/users/%s/chat", ts.URL, clientId)

	clientService.ReceiveMessages(ts.URL, cancel)
	assert.Equal(t, "Max: wubbalubbadubdub\n", output.String())

	go clientService.ReceiveMessages(ts.URL, cancel)
	cancel()

	//TODO
}

// func TestPostMessages(t *testing.T) {
// 	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		w.Write([]byte("Max: wubbalubbadubdub"))
// 	}))
// 	defer ts.Close()
// 	ts.URL = fmt.Sprintf("%s/users/%s/message", ts.URL, clientId)

// 	clientService.reader = bufio.NewReader(strings.NewReader("Max: wubbalubbadubdub" + "\n"))
// 	err := clientService.PostMessage(ts.URL)
// 	assert.Nil(t, err)

// 	clientService.reader = bufio.NewReader(strings.NewReader("quit" + "\n"))
// 	err = clientService.PostMessage(ts.URL)
// 	assert.Error(t, err)
// }
