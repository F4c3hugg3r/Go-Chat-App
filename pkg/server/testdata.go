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
	dummyClient   = &Client{
		Name:      name,
		ClientId:  clientId,
		clientCh:  make(chan Response, 100),
		Active:    false,
		authToken: authToken,
		lastSign:  time.Now(),
	}
	dummyClient2 = &Client{
		Name:      name2,
		ClientId:  clientId2,
		clientCh:  make(chan Response, 100),
		Active:    true,
		authToken: authToken2,
		lastSign:  time.Now(),
	}
	dummyClientInactive = &Client{
		Name:      name,
		ClientId:  clientId,
		clientCh:  make(chan Response, 100),
		Active:    false,
		authToken: authToken,
		lastSign:  time.Now().Add(-time.Hour),
	}

	dummyExamples = []dummyRequests{
		{
			method:   "GET",
			clientId: clientId,
		},
		{
			method:   "GET",
			clientId: "",
		},
		{
			method:   "POST",
			clientId: clientId,
			body:     name,
			token:    authToken,
		},
		{
			method:   "POST",
			clientId: "",
			body:     name,
		},
		{
			method:   "POST",
			clientId: clientId,
			body:     "",
		},
		{
			method:   "POST",
			clientId: clientId2,
			body:     strings.Repeat("s", (int(1<<20) + 1)),
		},
		{
			method:   "POST",
			clientId: clientId,
			token:    "",
		},
		{
			method:   "POST",
			clientId: clientId2,
			token:    authToken,
		},
		{
			method:   "POST",
			clientId: clientId,
			token:    authToken2,
		},
	}
)

type dummyRequests struct {
	method   string
	clientId string
	body     string
	token    string
}
