package main

import (
	"log"

	"github.com/serious-snow/govm/cmd"
)

func main() {
	if err := cmd.Run(); err != nil {
		log.Fatalln(err)
	}
}
