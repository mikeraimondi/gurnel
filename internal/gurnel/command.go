package gurnel

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"
	"text/template"
)

type subcommand interface {
	Run(io.Reader, io.Writer, []string, *Config) error
	Name() string
	ShortHelp() string
	LongHelp() string
	Flag() flag.FlagSet
}

func Do(r io.Reader, w io.Writer, conf *Config) error {
	flag.Usage = func() {
		printUsage(w, conf.subcommands)
	}
	flag.Parse()

	args := flag.Args()
	return run(r, w, args, conf)
}

func run(r io.Reader, w io.Writer, args []string, conf *Config) error {
	if len(args) < 1 {
		printUsage(w, conf.subcommands)
		return fmt.Errorf("no subcommand supplied. Did you mean 'gurnel start'?")
	}

	if args[0] == "help" {
		return help(w, conf.subcommands, args[1:])
	}

	for _, cmd := range conf.subcommands {
		if cmd.Name() != args[0] {
			continue
		}

		flagSet := cmd.Flag()
		name := cmd.Name()
		flagSet.Usage = func() {
			fmt.Fprintf(w, "usage: %s\n\n", name)
		}
		if err := flagSet.Parse(args[1:]); err != nil {
			return fmt.Errorf("parsing flags: %w", err)
		}
		args = flagSet.Args()
		if err := cmd.Run(r, w, args, conf); err != nil {
			return err
		}
		return nil
	}

	return fmt.Errorf(
		"unknown subcommand %q\n Run 'gurnel help' for usage",
		args[0],
	)
}

func printUsage(w io.Writer, commands []subcommand) {
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

func help(w io.Writer, commands []subcommand, args []string) error {
	if len(args) == 0 {
		printUsage(w, commands)
		return nil
	}
	if len(args) != 1 {
		return errors.New("too many arguments given")
	}

	arg := args[0]

	helpTemplate := "usage: gurnel {{.Name}}\n{{.LongHelp | trim}}"
	for _, cmd := range commands {
		if cmd.Name() == arg {
			tmpl(w, helpTemplate, cmd)
			return nil
		}
	}

	return errors.New("unknown help topic %#q.  Run 'gurnel help'")
}

func tmpl(w io.Writer, text string, data interface{}) {
	t := template.New("top")
	t.Funcs(template.FuncMap{"trim": strings.TrimSpace})
	template.Must(t.Parse(text))
	if err := t.Execute(w, data); err != nil {
		panic(err)
	}
}
