package gurnel

import (
	"bytes"
	"flag"
	"io"
	"testing"

	"github.com/mikeraimondi/gurnel/internal/test"
)

type testCmd struct {
	runFn  func(io.Reader, io.Writer, []string, *Config) error
	helpFn func() string
}

func (t *testCmd) Name() string       { return "testing" }
func (t *testCmd) ShortHelp() string  { return "" }
func (t *testCmd) Flag() flag.FlagSet { return flag.FlagSet{} }
func (t *testCmd) LongHelp() string   { return t.helpFn() }

func (t *testCmd) Run(r io.Reader, w io.Writer, args []string, conf *Config) error {
	return t.runFn(r, w, args, conf)
}

func TestRun(t *testing.T) {
	testCases := []struct {
		desc string
		args []string
		conf Config
		err  string
		out  []string
	}{
		{
			desc: "with no subcommand",
			args: []string{},
			conf: Config{},
			err:  "no subcommand",
			out:  []string{},
		},
		{
			desc: "with a nonexistent subcommand",
			args: []string{"foobar"},
			conf: Config{},
			err:  "unknown subcommand",
			out:  []string{},
		},
		{
			desc: "with the help subcommand",
			args: []string{"help"},
			conf: Config{},
			err:  "",
			out: []string{
				"usage",
				"commands are",
			},
		},
		{
			desc: "when invoking a subcommand",
			args: []string{"testing"},
			conf: Config{
				subcommands: []subcommand{
					&testCmd{
						runFn: func(r io.Reader, w io.Writer, _ []string, _ *Config) error {
							io.WriteString(w, "foo bar baz")
							return nil
						},
					},
				},
			},
			err: "",
			out: []string{"foo bar baz"},
		},
		{
			desc: "when invoking a subcommand's help",
			args: []string{"help", "testing"},
			conf: Config{
				subcommands: []subcommand{
					&testCmd{
						helpFn: func() string {
							return "help for testing"
						},
					},
				},
			},
			err: "",
			out: []string{"help for testing"},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			out := bytes.Buffer{}
			err := run(&bytes.Buffer{}, &out, tC.args, &tC.conf)

			test.CheckErr(t, tC.err, err)
			test.CheckOutput(t, tC.out, out.String())
		})
	}
}
