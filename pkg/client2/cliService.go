package client2

import (
	"fmt"
	"log"
	"strings"
)

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
