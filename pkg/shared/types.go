package shared

import (
	"errors"
)

// routes
const (
	PostPlugin = iota
	PostRegister
	Delete
	Get
	SignalWebRTC
)

const UnregisterFlag = "- Du bist nun vom Server getrennt -"
const RegisterFlag = "- Du bist registriert -"

const AddGroupFlag = "Add Group"
const LeaveGroupFlag = "Leave Group"

const ICECandidateFlag = "ICE Candidate"
const OfferSignalFlag = "Offer Signal"
const AnswerSignalFlag = "Answer Signal"
const RollbackSignalFlag = "Rollback Signal"
const StableSignalFlag = "Stable Flag"
const ConnectedFlag = "Connected"
const FailedConnectionFlag = "Connection Failed"

// const StartCallFlag = "Call Started"
const IgnoreResponseTag = "Ignore Response"

var (
	ErrNotAvailable   error = errors.New("item is not available")
	ErrNoPermission   error = errors.New("you have no permission")
	ErrEmptyString    error = errors.New("the string is empty")
	ErrTimeoutReached error = errors.New("timeout was reached")
	ErrChannelClosed  error = errors.New("access")
	ErrParsing        error = errors.New("the input couldn't be parsed")
)

// Message contains the name and id of the requester and the message (content) itsself
// as well as the uses plugin and groupId if a user is in a group
type Message struct {
	Name     string `json:"name"`
	Content  string `json:"content"`
	Plugin   string `json:"plugin"`
	ClientId string `json:"clientId"`
	GroupId  string `json:"groupId"`
}

// Response contains the name and id of the sender, the response (content) itsself
// and an error string
type Response struct {
	ClientId string `json:"clientId"`
	RspName  string `json:"name"`
	Content  string `json:"content"`
	Err      string `json:"errorString"`
}

// JsonGroup contains an id the groupname and the size of the group
// Notice that there is another Group struct in groupRegistry which
// has some extra fields which are used for server internal logic
type JsonGroup struct {
	GroupId string `json:"groupId"`
	Name    string `json:"name"`
	Size    int    `json:"size"`
}

type Connection struct {
	CallState string
	Connected bool
}

// logging
type Logg struct {
	Text   string
	Method string
}
