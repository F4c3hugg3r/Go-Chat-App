package server

// Client is a communication participant who has a name, unique id and
// channel to receive messages
type Client struct {
	name      string
	clientId  string
	clientCh  chan Message
	active    bool
	authToken string
}

// Message contains the name of the sender and the message (content) itsself
type Message struct {
	name    string
	content string
}
