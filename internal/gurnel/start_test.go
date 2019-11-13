package gurnel

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/mikeraimondi/gurnel/internal/test"
)

type testReader struct {
	t     *testing.T
	input []string
	i     int
	done  bool
}

func (tr *testReader) Read(p []byte) (int, error) {
	if tr.done || tr.i >= len(tr.input) {
		tr.done = false
		return 0, io.EOF
	}

	n, err := strings.NewReader(tr.input[tr.i]).Read(p)
	tr.i++
	tr.done = true
	return n, err
}

func TestStart(t *testing.T) {
	testCases := []struct {
		desc  string
		input string
		conf  Config
		err   string
		out   []string
	}{
		{
			desc:  "with input exceeding the minimum length",
			input: "foo bar baz",
			conf: Config{
				MinimumWordCount: 3,
			},
			out: []string{"begin entry preview", "foo bar baz", "exiting"},
		},
		{
			desc:  "with input less than the minimum length",
			input: "foo bar",
			conf: Config{
				MinimumWordCount: 3,
			},
			out: []string{"2 words", "Insufficient word count"},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			tC.conf.BeeminderEnabled = false
			_, filename, _, _ := runtime.Caller(0)
			dir := filepath.Dir(filepath.Dir(filepath.Dir(filename)))
			tC.conf.Editor = filepath.Join(dir, "test", "no_op_editor.sh")
			tC.conf.clock = &test.FixedClock{}

			dir, cleanup := test.SetupTestDir(t)
			defer cleanup()

			cmd := startCmd{}
			inReader := testReader{
				t:     t,
				input: []string{"1\n", "1\n", "1\n", "1\n", "n\n"},
			}
			out := bytes.Buffer{}
			errC := make(chan error)

			go func() {
				errC <- cmd.Run(&inReader, &out, []string{}, &tC.conf)
			}()

			files, _ := ioutil.ReadDir(dir)
			var file os.FileInfo
			for {
				if len(files) == 1 {
					file = files[0]
					break
				} else if len(files) > 1 {
					t.Fatalf("expected 1 file in directory. got %d", len(files))
				}
				files, _ = ioutil.ReadDir(dir)
			}

			defer test.WriteFile(t, filepath.Join(dir, file.Name()), tC.input)()

			err := <-errC
			test.CheckErr(t, tC.err, err)

			test.CheckOutput(t, tC.out, out.String())
		})
	}
}
