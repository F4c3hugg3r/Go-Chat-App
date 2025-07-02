package server

import (
	"strings"
	"time"
)

var (
	name          = "Arndt"
	name2         = "Len"
	clientId      = "clientId-DyGWNnLrLWnbuhf-LgBUAdAxdZf-U1pgRw"
	clientId2     = "clientId2-yGWNnLrLWnbuhf-LgBUAdAxdZf-U1pgRw"
	authToken     = "authId-5EDyGWNnLrLWnbuhf-LgBUAdAxdZf-U1pgRwc7ex1dt5EDyGWNnLrLWnbuhf-LgBUAdAxdZf-U1pgRw"
	authToken2    = "authId2-EDyGWNnLrLWnbuhf-LgBUAdAxdZf-U1pgRwc7ex1dt5EDyGWNnLrLWnbuhf-LgBUAdAxdZf-U1pgRw"
	dummyResponse = Response{Name: name, Content: "What's poppin"}
	dummyClient   = Client{
		Name:      name,
		ClientId:  clientId,
		clientCh:  make(chan *Response, 100),
		Active:    false,
		authToken: authToken,
		lastSign:  time.Now(),
	}
	dummyClient2 = Client{
		Name:      name2,
		ClientId:  clientId2,
		clientCh:  make(chan *Response, 100),
		Active:    true,
		authToken: authToken2,
		lastSign:  time.Now(),
	}
	dummyClientInactive = Client{
		Name:      name,
		ClientId:  clientId,
		clientCh:  make(chan *Response, 100),
		Active:    false,
		authToken: authToken,
		lastSign:  time.Now().Add(-time.Hour),
	}

	dummyExamples = []dummyRequests{
		{
			//valid
			method:   "GET",
			clientId: clientId,
		},
		{
			method: "GET",
			//empty
			clientId: "",
		},
		{
			//valid
			method:   "POST",
			clientId: clientId,
			message:  Message{Name: "Arndt", Content: "wubbalubbadubdub", Plugin: "/broadcast", ClientId: clientId},
			token:    authToken,
		},
		{
			method: "POST",
			//empty
			clientId: "",
			message:  Message{Name: "Arndt", Content: "wubbalubbadubdub", Plugin: "/broadcast", ClientId: clientId},
		},
		{
			method:   "POST",
			clientId: clientId,
			//empty content
			message: Message{Name: "Arndt", Content: "", Plugin: "/broadcast", ClientId: clientId},
		},
		{
			method:   "POST",
			clientId: clientId2,
			//too large
			message: Message{Name: "Arndt", Content: strings.Repeat("s", (int(1<<20) + 1)), Plugin: "/broadcast", ClientId: clientId},
		},
		{
			method:   "POST",
			clientId: clientId,
			message:  Message{Name: "Arndt", Content: "wubbalubbadubdub", Plugin: "/broadcast", ClientId: clientId},
			//empty
			token: "",
		},
		{
			//register plugin
			method:   "POST",
			clientId: clientId2,
			message:  Message{Name: "Arndt", Content: "wubbalubbadubdub", Plugin: "/register", ClientId: clientId2},
			token:    authToken,
		},
		{
			method:   "POST",
			clientId: clientId,
			//invalid plugin
			message: Message{Name: "Arndt", Content: "wubbalubbadubdub", Plugin: "/skdalskjd", ClientId: clientId},
			token:   authToken2,
		},
	}
)

type dummyRequests struct {
	method   string
	clientId string
	message  Message
	token    string
}
