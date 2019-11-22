package gurnel

import (
	"bytes"
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
			testClock := test.FixedClock{}

			for i, words := range tC.entryWords {
				if words == noentry {
					continue
				}

				entryTime := testClock.Now().Add(-time.Duration(24*i) * time.Hour)
				t.Log(entryTime)
				entry, err := NewEntry(dir, entryTime)
				if err != nil {
					t.Fatalf("saving entry: %s", err)
				}

				defer test.WriteFile(t, entry.Path, words)()
			}

			cmd := statsCmd{}
			out := bytes.Buffer{}
			conf := Config{clock: &testClock}
			if err := cmd.Run(&bytes.Buffer{}, &out, []string{}, &conf); err != nil {
				t.Fatalf("expected no error. got %s", err)
			}

			test.CheckOutput(t, tC.out, out.String())
		})
	}
}
