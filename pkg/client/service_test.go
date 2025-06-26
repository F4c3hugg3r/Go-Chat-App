package client

// import (
// 	"bufio"
// 	"bytes"
// 	"fmt"
// 	"io"
// 	"net/http"
// 	"net/http/httptest"
// 	"os"
// 	"strings"
// 	"testing"

// 	"github.com/stretchr/testify/assert"
// )

// var (
// 	clientService = &Client{
// 		clientId:   clientId,
// 		authToken:  authToken,
// 		reader:     bufio.NewReader(os.Stdin),
// 		writer:     io.Writer(os.Stdout),
// 		httpClient: &http.Client{},
// 	}
// 	name     = "Len"
// 	clientId = "clientId-DyGWNnLrLWnbuhf-LgBUAdAxdZf-U1pgRw"
// 	//clientId2  = "clientId2-yGWNnLrLWnbuhf-LgBUAdAxdZf-U1pgRw"
// 	authToken = "authId-5EDyGWNnLrLWnbuhf-LgBUAdAxdZf-U1pgRwc7ex1dt5EDyGWNnLrLWnbuhf-LgBUAdAxdZf-U1pgRw"
// 	//authToken2 = "authId2-EDyGWNnLrLWnbuhf-LgBUAdAxdZf-U1pgRwc7ex1dt5EDyGWNnLrLWnbuhf-LgBUAdAxdZf-U1pgRw"
// )

// // with testify
// func TestRegister(t *testing.T) {
// 	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		w.Write([]byte(authToken))
// 	}))
// 	defer ts.Close()
// 	clientService.authToken = ""

// 	err := clientService.Register(ts.URL)
// 	assert.Error(t, err)

// 	ts.URL = fmt.Sprintf("%s/users/%s", ts.URL, clientId)

// 	clientService.reader = bufio.NewReader(strings.NewReader(name + "\n"))
// 	err = clientService.Register(ts.URL)
// 	assert.NoError(t, err)
// 	assert.Equal(t, clientService.authToken, authToken)
// }

// func TestGetMessages(t *testing.T) {
// 	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		w.Write([]byte("Max: wubbalubbadubdub"))
// 	}))
// 	defer ts.Close()

// 	var output bytes.Buffer
// 	clientService.writer = &output

// 	ts.URL = fmt.Sprintf("%s/users/%s/chat", ts.URL, clientId)

// 	clientService.GetMessages(ts.URL)
// 	assert.Equal(t, "Max: wubbalubbadubdub", output.String())
// }

// func TestPostMessages(t *testing.T) {
// 	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		w.Write([]byte("Max: wubbalubbadubdub"))
// 	}))
// 	defer ts.Close()
// 	ts.URL = fmt.Sprintf("%s/users/%s/message", ts.URL, clientId)

// 	clientService.reader = bufio.NewReader(strings.NewReader("Max: wubbalubbadubdub" + "\n"))
// 	quit, err := clientService.PostMessage(ts.URL)
// 	assert.Nil(t, err)
// 	assert.Equal(t, 0, quit)

// 	clientService.reader = bufio.NewReader(strings.NewReader("quit" + "\n"))
// 	quit, err = clientService.PostMessage(ts.URL)
// 	assert.Error(t, err)
// 	assert.Equal(t, 1, quit)
// }
