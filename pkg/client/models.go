package client

import (
	"bufio"
	"io"
	"net/http"
)

type Client struct {
	clientId   string
	reader     *bufio.Reader
	writer     io.Writer
	authToken  string
	httpClient *http.Client
}

// Message contains the name of the sender and the message (content) itsself
type Message struct {
	Name    string `json:"name"`
	Content string `json:"content"`
	Plugin  string `json:"plugin"`
}
