package client2

import (
	"fmt"
	"log"
	"strings"

	"github.com/c-bata/go-prompt"
)

func NewUserService(c *ChatClient) *UserService {
	return &UserService{c: c}
}

func displayMessage(rsp *Response) {
	if strings.HasPrefix(rsp.Content, "[") {
		output, err := JSONToTable(rsp.Content)
		if err != nil {
			log.Printf("%v: Fehler beim Abrufen ist aufgetreten", err)

			return
		}

		fmt.Println(output)
		return
	}

	responseString := fmt.Sprintf("%s: %s\n", rsp.Name, rsp.Content)
	fmt.Println(responseString)
}

func (u *UserService) parseInputToMessage(input string) (*Message, error) {

	//TODO string mit regex einlesen und message zurück geben
	if u.c.Registered == false {
		return nil, fmt.Errorf("%w: you are not registered yet", ErrNotRegistered)
	}

	// switch {
	// case strings.HasPrefix(input, "/register"):
	// 	{
	// 		clientName, _ := strings.CutPrefix(input, "/register ")
	// 		u.c.PollMessages()
	// 		//printen
	// 		u.c.SendMessage(&Message{Name: clientName, Plugin: "/register", ClientId: u.c.clientId})
	// 		return
	// 	}
	// //nil pointer reference wenn noch nicht registriert
	// case !strings.HasPrefix(input, "/"):
	// 	{
	// 		if u.c.Registered == false {
	// 			log.Printf("you are not registered yet")
	// 			return
	// 		}
	// 		u.c.PollMessages()
	// 		//printen
	// 		u.c.SendMessage(&Message{Name: u.c.clientName, Plugin: "/broadcast", Content: input, ClientId: u.c.clientId})
	// 		return
	// 	}
	// case strings.HasPrefix(input, "/private"):
	// 	{
	// 		if u.c.Registered == false {
	// 			log.Printf("you are not registered yet")
	// 			return
	// 		}
	// 		opposingClientId := strings.Fields(input)[1]
	// 		message, _ := strings.CutPrefix(input, fmt.Sprintf("/private %s ", opposingClientId))

	// 		u.c.PollMessages()
	// 		//printen
	// 		u.c.SendMessage(&Message{Name: u.c.clientName, Plugin: "/private", ClientId: opposingClientId, Content: message})
	// 		return
	// 	}
	// default:
	// 	{
	// 		if u.c.Registered == false {
	// 			log.Printf("you are not registered yet")
	// 			return
	// 		}
	// 		plugin := strings.Fields(input)[0]

	// 		content := strings.ReplaceAll(input, plugin, "")
	// 		content, _ = strings.CutPrefix(content, " ")

	// 		u.c.PollMessages()
	// 		//printen
	// 		u.c.SendMessage(&Message{Name: u.c.clientName, Plugin: plugin, Content: content, ClientId: u.c.clientId})
	// 	}
	// }
	return nil, nil
}

func (u *UserService) Executor(input string) {
	input = strings.TrimSuffix(input, "\n")

	// parsing function fmt regex
	msg, err := u.parseInputToMessage(input)
	if err != nil {
		log.Printf("%v: wrong input", err)
	}

	err = u.plugins.FindAndExecute(msg)()
	if err != nil {
		log.Printf("%v: couldn't send message", err)
	}
}

func (u *UserService) Completer(d prompt.Document) []prompt.Suggest {
	//durch list plugins iterieren

	s := []prompt.Suggest{
		{Text: "/help", Description: "Zeigt die Hilfe an"},
		{Text: "/quit", Description: "Beendet das Programm"},
		{Text: "/private", Description: "Sendet eine private message"},
		{Text: "/users", Description: "Listet alle User"},
		{Text: "/time", Description: "Zeigt die aktuelle Zeit an"},
		{Text: "/register", Description: "Registriert dich beim Server"},
	}

	//TODO bei private weiteren vorschlag für clientids

	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
}
