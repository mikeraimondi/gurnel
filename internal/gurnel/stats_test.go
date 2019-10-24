package gurnel

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mikeraimondi/gurnel/internal/test"
)

func TestStats(t *testing.T) {
	const noentry = "NOENTRY"

	testCases := []struct {
		desc       string
		entryWords []string
		out        []string
	}{
		{
			desc:       "with no entries",
			entryWords: []string{},
			out:        []string{"no entries found"},
		},
		{
			desc:       "with an empty entry",
			entryWords: []string{""},
			out: []string{
				"word count: 0",
				`100.00% of days`,
			},
		},
		{
			desc:       "with one populated entry",
			entryWords: []string{"foo bar baz"},
			out: []string{
				"word count: 3",
				`100.00% of days`,
			},
		},
		{
			desc:       "with two populated entries",
			entryWords: []string{"foo bar", "baz"},
			out: []string{
				"word count: 3",
				`100.00% of days`,
			},
		},
		{
			desc:       "with one entry and one day missed",
			entryWords: []string{noentry, "baz"},
			out: []string{
				"word count: 1",
				`50.00% of days`,
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			dir, cleanup := test.SetupTestDir(t)
			defer cleanup()

			for i, words := range tC.entryWords {
				if words == noentry {
					continue
				}

				entryTime := time.Now().Add(-time.Duration(24*i) * time.Hour)
				p := &Entry{Path: dir + string(filepath.Separator) + entryTime.Format(entryFormat)}
				err := p.Save()
				if err != nil {
					t.Fatalf("saving entry: %s", err)
				}

				defer test.WriteFile(t, p.Path, words)()
			}

			cmd := statsCmd{}
			out := bytes.Buffer{}
			if err := cmd.Run(&bytes.Buffer{}, &out, []string{}, &config{}); err != nil {
				t.Fatalf("expected no error. got %s", err)
			}

			for _, expectedOut := range tC.out {
				if !strings.Contains(strings.ToLower(out.String()), strings.ToLower(expectedOut)) {
					t.Fatalf("expected output containing %s. got %q", expectedOut, out.String())
				}
			}
		})
	}
}
