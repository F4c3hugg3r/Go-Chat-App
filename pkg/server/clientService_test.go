package server

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestReceive(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	client := Client{
		Name:      name,
		ClientId:  clientId,
		clientCh:  make(chan *Response, 100),
		Active:    false,
		authToken: authToken,
		lastSign:  time.Now().UTC().Add(-time.Hour),
	}

	client.clientCh <- &dummyResponse

	rsp, err := client.Receive(ctx)
	assert.Nil(t, err)

	go func() {
		rsp, err = client.Receive(ctx)
	}()
	cancel()
	assert.Nil(t, rsp)
	assert.ErrorIs(t, err, TimeoutReachedError)
}

func TestSend(t *testing.T) {
	//TODO
}
