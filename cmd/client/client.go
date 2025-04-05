package main

import (
	"bufio"
	flera "flera/pkg/client"
	"fmt"
	"os"
	"strings"
)

var id uint32

func main() {
	args := os.Args[1:]

	for i, arg := range args {
		fmt.Printf("Argument %d: %s\n", i+1, arg)
	}

	if len(args) < 1 {
		os.Exit(1)
	}

	c := flera.New()
	c.Connect(args[0])

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("-> ")
		input, err := reader.ReadString('\n')
		if err != nil {
			os.Exit(2)
		}
		input = strings.Replace(input, "\n", "", -1)

		if err := c.Send(0, []byte(input)); err != nil {
			os.Exit(3)
		}

	}
}
