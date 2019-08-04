package gurnel

import (
	"bytes"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func TestStart(t *testing.T) {
	testCases := []struct {
		desc  string
		input string
		conf  config
		err   string
		out   []string
	}{
		{
			desc:  "with input exceeding the minimum length",
			input: "foo bar baz",
			conf: config{
				MinimumWordCount: 3,
			},
			out: []string{"begin entry preview", "foo bar baz"},
		},
		{
			desc:  "with input less than the minimum length",
			input: "foo bar",
			conf: config{
				MinimumWordCount: 3,
			},
			out: []string{"2 words", "Insufficient word count"},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			tC.conf.BeeminderEnabled = false
			tC.conf.Editor = "ed"
			dir, err := ioutil.TempDir("", "gurnel_test")
			if err != nil {
				t.Fatal(err)
			}
			if err = os.Chdir(dir); err != nil {
				t.Fatal(err)
			}
			cmd := startCmd{}
			in := &bytes.Buffer{}
			in.WriteString("a\n" + tC.input + "\n" + ".\n" + "w\n" + "q\n" + "1\n" + "1\n" + "1\n" + "n\n")
			out := bytes.Buffer{}
			errC := make(chan error)
			go func() {
				errC <- cmd.Run(in, &out, []string{}, &tC.conf)
			}()

			err = <-errC
			if tC.err == "" {
				if err != nil {
					t.Fatalf("expected no error. got %s", err)
				}
			} else {
				if !strings.Contains(err.Error(), tC.err) {
					t.Fatalf("expected an error containing %s. got %s", tC.err, err)
				}
			}
			for _, expectedOut := range tC.out {
				if !strings.Contains(out.String(), expectedOut) {
					t.Fatalf("expected output containing %s. got %q", tC.out, out.String())
				}
			}
		})
	}
}
