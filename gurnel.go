package main

import (
	"fmt"
	"os"
	"sort"
)

// TODO make configurable
const (
	minWordCount = 750
	entryFormat  = "2006-01-02-Journal-Entry-for-Jan-2" + ".md"
	entryRegex   = `\d{4}-\d{2}-\d{2}-Journal-Entry-for-\D{3}-\d{1,2}` + ".md"
)

var (
	commandHandlers map[string]gurnelCmd
	commands        []string
)

type gurnelCmd struct {
	f             func([]string) error
	condensedHelp string
	fullHelp      string
}

func init() {
	commandHandlers = make(map[string]gurnelCmd)
}

func main() {
	registerCommandHandler("root", root)
	registerCommandHandler("start", startCmd)
	registerCommandHandler("stats", statsCmd)
	// TODO add "init" command. It should check that the directory is empty, create a config file, and do a 'git init'
	executeCommand()
}

func registerCommandHandler(command string, f func() gurnelCmd) {
	commandHandlers[command] = f()
	if command != "root" {
		commands = append(commands, command)
	}
}

func executeCommand() {
	var args []string
	l := len(os.Args)
	if l == 1 {
		args = []string{"root"}
	} else if l > 1 {
		args = os.Args[1:]
	} else {
		fmt.Println("Internal error")
		return
	}
	cmd, switches := parseOpts(args)
	if val, ok := commandHandlers[cmd]; ok {
		if err := val.f(switches); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	} else {
		fmt.Println("Unrecognized command")
		displayHelpToc()
	}
}

func parseOpts(args []string) (command string, switches []string) {
	return args[0], args[1:]
}

func displayHelpToc() {
	fmt.Print("usage: gurnel <command>\n\n")
	fmt.Println("Commands:")
	maxCmdLength := 0
	for _, cmd := range commands {
		l := len(cmd)
		if l > maxCmdLength {
			maxCmdLength = l
		}
	}
	sort.Strings(commands)
	for _, cmd := range commands {
		fmt.Print(cmd)
		for i := 0; i < (maxCmdLength-len(cmd))+2; i++ {
			fmt.Print(" ")
		}
		fmt.Println(commandHandlers[cmd].condensedHelp)
	}
}
