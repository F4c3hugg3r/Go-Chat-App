package server

import (
	"errors"
	"time"
)

var (
	ClientNotAvailableError error = errors.New("client is not available")
	NoPermissionError       error = errors.New("you have no permission")
	EmptyStringError        error = errors.New("the string is empty")
)

// Client is a communication participant who has a name, unique id and
// channel to receive messages
type Client struct {
	Name      string
	ClientId  string
	clientCh  chan Response
	Active    bool
	authToken string
	lastSign  time.Time
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
