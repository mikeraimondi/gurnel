package main

import (
	"fmt"
	"os"
)

var commandHandlers map[string]func() error

func init() {
	commandHandlers = make(map[string]func() error)
}

func main() {
	registerCommandHandler("start", start)
	registerCommandHandler("percent", percent)
	executeCommand()
}

func registerCommandHandler(command string, f func() error) {
	commandHandlers[command] = f
	// TODO dynamically add help text
}

func executeCommand() {
	if len(os.Args) > 1 {
		if val, ok := commandHandlers[os.Args[1]]; ok {
			if err := val(); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		} else {
			fmt.Println("Unrecognized command")
			help()
		}
	} else {
		help()
	}
}

func help() {
	fmt.Println("Commands:")
	fmt.Println("  start    Begin journal entry for today")
	fmt.Println("  percent  View percentage of days journaled")
}
