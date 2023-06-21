package main

import (
	"log"
	"os"

	"git.epam.com/vadym_ulitin/lets-go-chat/api/server"
)

func main() {
	var pass string
	if len(os.Args[1:]) == 1 {
		pass = os.Args[1]
	}
	log.Fatal(server.Run(pass))
}
