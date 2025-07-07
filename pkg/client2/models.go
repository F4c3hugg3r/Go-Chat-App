package client2

import (
	"errors"
	"net/http"
	"sync"
)

var (
	ErrNoPermission error = errors.New("you have no permission")
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
	url        string
}

type UserService struct {
	chatClient *ChatClient
	plugins    *PluginRegistry
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

type Plugin struct {
	Command     string
	Description string
}
