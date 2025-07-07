package client2

import (
	"fmt"
	"log"
	"slices"
	"strings"

	"github.com/c-bata/go-prompt"
)

func NewUserService(c *ChatClient) *UserService {
	return &UserService{
		chatClient: c,
		plugins:    RegisterPlugins(c),
	}
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
	//displayed nicht richtig, liegt evtl an prompt package
	fmt.Println(responseString)
	return nil
}

func (u *UserService) parseInputToMessage(input string) (*Message, error) {
	input = strings.TrimSuffix(input, "\n")
	var plugin string

	ok := strings.HasPrefix(input, "/")
	switch ok {
	case true:
		plugin = strings.Fields(input)[0]
	case false:
		plugin = "/broadcast"
	}

	content := strings.ReplaceAll(input, plugin, "")
	content, _ = strings.CutPrefix(content, " ")

	return u.chatClient.CreateMessage("", plugin, content, ""), nil
}

func (u *UserService) Executor(input string) {

	msg, err := u.parseInputToMessage(input)
	if err != nil {
		log.Printf("%v: wrong input", err)
	}

	responses := u.chatClient.PollMessages()
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

	//TODO nur Vorschläge beim ersten Wort

	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
}
