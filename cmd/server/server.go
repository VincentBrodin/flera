package main

import (
	flera "flera/pkg/server"
	"fmt"
	"os"
	"sync"
)

var connMap sync.Map

func main() {
	args := os.Args[1:]
	if len(args) < 1 {
		os.Exit(1)
	}

	s := flera.New()

	s.Register(0, MessageHandler)

	s.Start(args[0])
}

func MessageHandler(id uint32, data []byte) error {
	fmt.Printf("%d: %s\n", id, data)
	return nil
}
