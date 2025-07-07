package client2

import (
	"fmt"
	"log"
	"slices"
	"strings"

	"github.com/c-bata/go-prompt"
)

func NewUserService(c *ChatClient) *UserService {
	return &UserService{c: c}
}

func displayResponse(rsp *Response) error {
	if strings.HasPrefix(rsp.Content, "[") {
		output, err := JSONToTable(rsp.Content)
		if err != nil {
			return fmt.Errorf("%v: error formatting json to table", err)
		}

		fmt.Println(output)
		return nil
	}

	responseString := fmt.Sprintf("%s: %s\n", rsp.Name, rsp.Content)
	fmt.Println(responseString)
	return nil
}

func (u *UserService) parseInputToMessage(input string) (*Message, error) {
	//TODO string - einheitlich - mit fmt regex einlesen und message zurück geben
	input = strings.TrimSuffix(input, "\n")

	if strings.HasPrefix(input, "/register") {
		clientName := strings.Fields(input)[1]
		return &Message{Name: clientName, Plugin: "/register", Content: clientName, ClientId: u.c.clientId}, nil
	}

	if !strings.HasPrefix(input, "/") {
		return &Message{Name: u.c.clientName, Plugin: "/broadcast", Content: input, ClientId: u.c.clientId}, nil
	}

	if strings.HasPrefix(input, "/private") {
		opposingClientId := strings.Fields(input)[1]
		message, _ := strings.CutPrefix(input, fmt.Sprintf("/private %s ", opposingClientId))

		return &Message{Name: u.c.clientName, Plugin: "/private", ClientId: opposingClientId, Content: message}, nil
	}

	plugin := strings.Fields(input)[0]

	content := strings.ReplaceAll(input, plugin, "")
	content, _ = strings.CutPrefix(content, " ")

	return &Message{Name: u.c.clientName, Plugin: plugin, Content: content, ClientId: u.c.clientId}, nil

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

	msg, err := u.parseInputToMessage(input)
	if err != nil {
		log.Printf("%v: wrong input", err)
	}

	responses := u.c.PollMessages()
	for _, rsp := range responses {
		err = displayResponse(rsp)
		if err != nil {
			log.Printf("%v: response from %s couldn't be displayed", err, rsp.Name)
		}
	}

	err = u.plugins.FindAndExecute(msg)()
	if err != nil {
		log.Printf("%v: couldn't send message", err)
	}
}

func (u *UserService) Completer(d prompt.Document) []prompt.Suggest {
	s := []prompt.Suggest{}

	for command, plugin := range u.plugins.plugins {
		if !slices.Contains(u.plugins.invisible, command) {
			s = append(s, prompt.Suggest{Text: command, Description: plugin.Description()})
		}
	}

	//TODO bei private weiteren vorschlag für clientids

	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
}
