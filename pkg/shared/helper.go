package shared

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"strings"
)

// generateSecureToken generates a token containing random chars
func GenerateSecureToken(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}

	return base64.RawURLEncoding.EncodeToString(b)
}

// DecodeToResponse decodes a responseBody to a Response struct
func DecodeToResponse(body []byte) (*Response, error) {
	response := &Response{}
	dec := json.NewDecoder(strings.NewReader(string(body)))

	err := dec.Decode(&response)
	if err != nil {
		return response, err
	}

	return response, nil
}

// DecodeToMessage decodes a responseBody to a Message struct
func DecodeToMessage(body []byte) (Message, error) {
	message := Message{}
	dec := json.NewDecoder(strings.NewReader(string(body)))
	err := dec.Decode(&message)

	if err != nil {
		return message, err
	}

	return message, nil
}

// DecodeToGroup decodes a responseBody to a Group struct
func DecodeToGroup(body []byte) (*JsonGroup, error) {
	group := &JsonGroup{}
	dec := json.NewDecoder(strings.NewReader(string(body)))

	err := dec.Decode(&group)
	if err != nil {
		return group, err
	}

	return group, nil
}

func DecodeStringToJsonGroup(jsonGroup string) (*JsonGroup, error) {
	var group *JsonGroup
	dec := json.NewDecoder(strings.NewReader(jsonGroup))
	err := dec.Decode(&group)

	return group, err
}
