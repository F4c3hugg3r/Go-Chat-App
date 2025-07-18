package types

import (
	"errors"
)

var (
	ErrNoPermission error = errors.New("you have no permission")
	ErrParsing      error = errors.New("there was an errror while parsing your input")
)

const (
	PostPlugin = iota
	PostRegister
	Delete
	Get
)

const UnregisterFlag = "- Du bist nun vom Server getrennt -"
const RegisterFlag = "- Du bist registriert -"
const AddGroupFlag = "Add Group"
const LeaveGroupFlag = "Leave Group"

// Message contains the name of the requester and the message (content) itsself
type Message struct {
	Name     string `json:"name"`
	Content  string `json:"content"`
	Plugin   string `json:"plugin"`
	ClientId string `json:"clientId"`
	GroupId  string `json:"groupId"`
}

// Response contains the name of the sender and the response (content) itsself
type Response struct {
	Name    string `json:"name"`
	Content string `json:"content"`
	Err     string `json:"errorString"`
}

// Group contains
type Group struct {
	GroupId string `json:"groupId"`
	Name    string `json:"name"`
	Size    int    `json:"size"`
}
