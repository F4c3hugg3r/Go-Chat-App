package legacy

// import (
// 	"strings"
// 	"time"

// 	ty "github.com/F4c3hugg3r/Go-Chat-Server/pkg/shared"
// )

// var (
// 	ClientName    = "Arndt"
// 	ClientName2   = "Len"
// 	ClientId      = "clientId-DyGWNnLrLWnbuhf-LgBUAdAxdZf-U1pgRw"
// 	ClientId2     = "clientId2-yGWNnLrLWnbuhf-LgBUAdAxdZf-U1pgRw"
// 	AuthToken     = "authId-5EDyGWNnLrLWnbuhf-LgBUAdAxdZf-U1pgRwc7ex1dt5EDyGWNnLrLWnbuhf-LgBUAdAxdZf-U1pgRw"
// 	AuthToken2    = "authId2-EDyGWNnLrLWnbuhf-LgBUAdAxdZf-U1pgRwc7ex1dt5EDyGWNnLrLWnbuhf-LgBUAdAxdZf-U1pgRw"
// 	DummyResponse = ty.Response{RspName: ClientName, Content: "What's poppin"}
// 	DummyClient   = Client{
// 		Name:      ClientName,
// 		ClientId:  ClientId,
// 		clientCh:  make(chan *ty.Response, 100),
// 		GroupName: "",
// 		active:    false,
// 		authToken: AuthToken,
// 		lastSign:  time.Now(),
// 	}
// 	// dummyClient2 = Client{
// 	// 	ClientName:      Clientname2,
// 	// 	ClientId:  clientId2,
// 	// 	clientCh:  make(chan *Response, 100),
// 	// 	Active:    true,
// 	// 	authToken: authToken2,
// 	// 	lastSign:  time.Now(),
// 	// }
// 	// dummyClientInactive = Client{
// 	// 	ClientName:      Clientname,
// 	// 	ClientId:  clientId,
// 	// 	clientCh:  make(chan *Response, 100),
// 	// 	Active:    false,
// 	// 	authToken: authToken,
// 	// 	lastSign:  time.Now().Add(-time.Hour),
// 	// }

// 	DummyExamples = []dummyRequests{
// 		{
// 			//valid
// 			Method:   "GET",
// 			ClientId: ClientId,
// 		},
// 		{
// 			Method: "GET",
// 			//empty
// 			ClientId: "",
// 		},
// 		{
// 			//valid
// 			Method:   "POST",
// 			ClientId: ClientId,
// 			Message:  ty.Message{Name: "Arndt", Content: "wubbalubbadubdub", Plugin: "/broadcast", ClientId: ClientId},
// 			Token:    AuthToken,
// 		},
// 		{
// 			Method: "POST",
// 			//empty
// 			ClientId: "",
// 			Message:  ty.Message{Name: "Arndt", Content: "wubbalubbadubdub", Plugin: "/broadcast", ClientId: ClientId},
// 		},
// 		{
// 			Method:   "POST",
// 			ClientId: ClientId,
// 			//empty content
// 			Message: ty.Message{Name: "Arndt", Content: "", Plugin: "/broadcast", ClientId: ClientId},
// 		},
// 		{
// 			Method:   "POST",
// 			ClientId: ClientId2,
// 			//too large
// 			Message: ty.Message{Name: "Arndt", Content: strings.Repeat("s", (int(1<<20) + 1)), Plugin: "/broadcast", ClientId: ClientId},
// 		},
// 		{
// 			Method:   "POST",
// 			ClientId: ClientId,
// 			Message:  ty.Message{Name: "Arndt", Content: "wubbalubbadubdub", Plugin: "/broadcast", ClientId: ClientId},
// 			//empty
// 			Token: "",
// 		},
// 		{
// 			//register plugin
// 			Method:   "POST",
// 			ClientId: ClientId2,
// 			Message:  ty.Message{Name: "Arndt", Content: "wubbalubbadubdub", Plugin: "/register", ClientId: ClientId2},
// 			Token:    AuthToken,
// 		},
// 		{
// 			Method:   "POST",
// 			ClientId: ClientId,
// 			//invalid plugin
// 			Message: ty.Message{Name: "Arndt", Content: "wubbalubbadubdub", Plugin: "/skdalskjd", ClientId: ClientId},
// 			Token:   AuthToken2,
// 		},
// 	}
// )

// type dummyRequests struct {
// 	Method   string
// 	ClientId string
// 	Message  ty.Message
// 	Token    string
// }
