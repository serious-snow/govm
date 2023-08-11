package main

import (
	"github.com/serious-snow/govm/cmd"
	"log"
)

func main() {
	if err := cmd.Run(); err != nil {
		log.Fatalln(err)
	}
}
