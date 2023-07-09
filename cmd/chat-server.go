package main

import (
	"log"
	"os"

	"github.com/uilki/lgc/api/server"
	"github.com/uilki/lgc/api/wired"
)

func main() {
	var pass string
	if len(os.Args[1:]) == 1 {
		pass = os.Args[1]
	}
	s, err := wired.InitializeServer(pass)
	if err != nil {
		panic(err)
	}

	log.Fatal(server.Run(s))
}
