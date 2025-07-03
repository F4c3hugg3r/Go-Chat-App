package client2

import (
	"net/http"
	"sync"
)

type Reader interface {
	ReadString(delim byte) (string, error)
}

type ChatClient struct {
	clientName string
	clientId   string
	authToken  string
	HttpClient *http.Client
	Registered bool
	mu         *sync.Mutex
	Cond       *sync.Cond
	Output     chan *Response
	Input      chan *Message
	url        string
}

// Message contains the name of the requester and the message (content) itsself
type Message struct {
	Name     string `json:"name"`
	Content  string `json:"content"`
	Plugin   string `json:"plugin"`
	ClientId string `json:"clientId"`
}

// Response contains the name of the sender and the response (content) itsself
type Response struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}
