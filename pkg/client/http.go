package client

import (
	"bytes"
	"fmt"
	"net/http"
)

// GetRequest sends a GET Request to the server including the authorization token
func (c *Client) GetRequest(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: Fehler beim erstellen der GET request: ", err)
	}

	req.Header.Add("Authorization", c.authToken)

	res, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: Fehler beim senden der GET request: ", err)
	}
	return res, nil
}

// DeleteRequest sends a DELETE Request to delete the client out of the server
// including the authorization token
func (c *Client) DeleteRequest(url string, body []byte) (*http.Response, error) {
	parameteredUrl := fmt.Sprintf("%s/users/%s", url, c.clientId)
	req, err := http.NewRequest("DELETE", parameteredUrl, bytes.NewReader(body))

	if err != nil {
		return nil, fmt.Errorf("%w: Fehler beim Erstellen der DELETE req", err)

	}

	req.Header.Add("Authorization", c.authToken)
	req.Header.Add("Content-Type", "application/json")

	res, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: Fehler beim Absenden des Deletes", err)
	}

	return res, nil
}

// PostReqeust sends a Post Request to send a message to the server
// including the authorization token
func (c *Client) PostRequest(url string, body []byte) (*http.Response, error) {
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("%w: Fehler beim Erstellen der POST req", err)
	}

	req.Header.Add("Authorization", c.authToken)
	req.Header.Add("Content-Type", "application/json")

	res, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: Fehler beim Absenden der Nachricht", err)

	}

	return res, nil
}
