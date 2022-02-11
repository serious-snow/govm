package main

import (
	"govm/cmd"
	"log"
)

func main() {
	if err := cmd.Run(); err != nil {
		log.Fatalln(err)
	}
}
