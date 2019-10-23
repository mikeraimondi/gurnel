package gurnel

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"
)

type subcommand interface {
	Run(io.Reader, io.Writer, []string, *config) error
	Name() string
	ShortHelp() string
	LongHelp() string
	Flag() flag.FlagSet
}

// Do executes the program
func Do() error {
	var conf config
	if err := conf.load("gurnel", "gurnel.json"); err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	flag.Usage = func() {
		printUsage(conf.subcommands, os.Stderr)
	}
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		printUsage(conf.subcommands, os.Stderr)
		return fmt.Errorf("no subcommand supplied. Did you mean 'gurnel start'?")
	}

	if args[0] == "help" {
		help(conf.subcommands, args[1:])
		return nil
	}

	for _, cmd := range conf.subcommands {
		if cmd.Name() != args[0] {
			continue
		}

		flagSet := cmd.Flag()
		name := cmd.Name()
		flagSet.Usage = func() {
			fmt.Fprintf(os.Stderr, "usage: %s\n\n", name)
		}
		if err := flagSet.Parse(args[1:]); err != nil {
			return fmt.Errorf("parsing flags: %w", err)
		}
		args = flagSet.Args()
		if err := cmd.Run(os.Stdin, os.Stdout, args, &conf); err != nil {
			return err
		}
		return nil
	}

	return fmt.Errorf(
		"unknown subcommand %q\n Run 'gurnel help' for usage",
		args[0],
	)
}

func printUsage(commands []subcommand, w io.Writer) {
	bw := bufio.NewWriter(w)
	usageTemplate := `Gurnel is a simple journal manager.

Usage:

	gurnel command [arguments]

The commands are:
{{range .}}
  {{.Name | printf "%-11s"}} {{.ShortHelp}}{{end}}
Use "gurnel help [command]" for more information about a command.
`
	tmpl(bw, usageTemplate, commands)
	bw.Flush()
}

func help(commands []subcommand, args []string) {
	if len(args) == 0 {
		printUsage(commands, os.Stdout)
		return
	}
	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "usage: gurnel help command\n\nToo many arguments given.\n")
		os.Exit(2)
	}

	arg := args[0]

	helpTemplate := "usage: gurnel {{.Name}}\n{{.LongHelp | trim}}"
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
