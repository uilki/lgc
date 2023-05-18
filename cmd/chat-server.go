package main

import (
	"log"

	"git.epam.com/vadym_ulitin/lets-go-chat/api/server"
)

func main() {
	log.Fatal(server.Run())
}
