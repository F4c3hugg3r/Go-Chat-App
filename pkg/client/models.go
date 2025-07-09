package client

import (
	"errors"
	"net/http"
	"sync"
)

var (
	ErrNoPermission error = errors.New("you have no permission")
	ErrParsing      error = errors.New("there was an errror while parsing your input")
)

const (
	postPlugin = iota
	postRegister
	delete
	get
)

// ChatClient handles all network tasks
type ChatClient struct {
	clientName string
	clientId   string
	authToken  string
	HttpClient *http.Client
	Registered bool
	mu         *sync.RWMutex
	Cond       *sync.Cond
	Output     chan *Response
	url        string
	Endpoints  map[int]string
}

// UserService handles user inputs and outputs
type UserService struct {
	chatClient *ChatClient
	plugins    *PluginRegistry
	poll       bool
	typing     bool
	mu         *sync.RWMutex
	Cond       *sync.Cond
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
	Err     error  `json:"-"`
}
