package types

import (
	"errors"
)

var (
	ErrClientNotAvailable error = errors.New("client is not available")
	ErrNoPermission       error = errors.New("you have no permission")
	ErrEmptyString        error = errors.New("the string is empty")
	ErrTimeoutReached     error = errors.New("timeout was reached")
	ErrChannelClosed      error = errors.New("access")
)

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
