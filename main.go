package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"
)

// TODO make configurable
const (
	minWordCount = 750
)

type command struct {
	Run       func(cmd *command, args []string) error
	UsageLine string
	ShortHelp string
	LongHelp  string
	Flag      flag.FlagSet
}

func (c *command) Name() string {
	name := c.UsageLine
	i := strings.Index(name, " ")
	if i >= 0 {
		name = name[:i]
	}
	return name
}

func (c *command) usage() {
	fmt.Fprintf(os.Stderr, "usage: %s\n\n", c.UsageLine)
	os.Exit(2)
}

var commands = []*command{
	cmdStart,
	cmdStats,
}

func main() {
	flag.Usage = usage
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		usage()
	}

	if args[0] == "help" {
		help(args[1:])
		return
	}

	for _, cmd := range commands {
		if cmd.Name() == args[0] && cmd.Run != nil {
			cmd.Flag.Usage = func() { cmd.usage() }
			cmd.Flag.Parse(args[1:])
			args = cmd.Flag.Args()
			if err := cmd.Run(cmd, args); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v", err)
				os.Exit(2)
			}
			return
		}
	}

	fmt.Fprintf(os.Stderr, "gurnel: unknown subcommand %q\n Run 'gurnel help' for usage.\n", args[0])
	os.Exit(2)
}

var usageTemplate = `Gurnel is a simple journal manager.

Usage:

	gurnel command [arguments]

The commands are:
{{range .}}
  {{.Name | printf "%-11s"}} {{.ShortHelp}}{{end}}
Use "gurnel help [command]" for more information about a command.
`

var helpTemplate = `usage: gurnel {{.UsageLine}}
{{.LongHelp | trim}}
`

func printUsage(w io.Writer) {
	bw := bufio.NewWriter(w)
	tmpl(bw, usageTemplate, commands)
	bw.Flush()
}

func usage() {
	printUsage(os.Stderr)
	os.Exit(2)
}

func help(args []string) {
	if len(args) == 0 {
		printUsage(os.Stdout)
		return
	}
	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "usage: gurnel help command\n\nToo many arguments given.\n")
		os.Exit(2)
	}

	arg := args[0]

	for _, cmd := range commands {
		if cmd.Name() == arg {
			tmpl(os.Stdout, helpTemplate, cmd)
			return
		}
	}

	fmt.Fprintf(os.Stderr, "Unknown help topic %#q.  Run 'gurnel help'.\n", arg)
	os.Exit(2)
}

func tmpl(w io.Writer, text string, data interface{}) {
	t := template.New("top")
	t.Funcs(template.FuncMap{"trim": strings.TrimSpace})
	template.Must(t.Parse(text))
	if err := t.Execute(w, data); err != nil {
		panic(err)
	}
}
