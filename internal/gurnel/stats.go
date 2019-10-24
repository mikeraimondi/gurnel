package gurnel

import (
	"bytes"
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"text/tabwriter"
	"time"

	"github.com/mikeraimondi/gurnel/internal/bindata"
)

type statsCmd struct{}

func (*statsCmd) Name() string       { return "stats" }
func (*statsCmd) ShortHelp() string  { return "View journal statistics" }
func (*statsCmd) Flag() flag.FlagSet { return flag.FlagSet{} }

func (*statsCmd) LongHelp() string {
	return "Unusually frequent/infrequent words are relative " +
		"to a Google Ngram corpus of scanned literature"
}

func (*statsCmd) Run(_ io.Reader, w io.Writer, args []string, conf *config) error {
	refFreqsCSV, err := bindata.Asset("eng-us-10000-1960.csv")
	if err != nil {
		return fmt.Errorf("loading asset: %s", err)
	}
	csvReader := csv.NewReader(bytes.NewReader(refFreqsCSV))
	csvReader.FieldsPerRecord = 2

	refFreqs := make(map[string]float64)
	for {
		record, csvErr := csvReader.Read()
		if csvErr == io.EOF {
			break
		}
		if csvErr != nil {
			return csvErr
		}

		if record[0] == "" || record[1] == "" {
			return fmt.Errorf("invalid input")
		}
		freq, csvErr := strconv.ParseFloat(record[1], 64)
		if csvErr != nil {
			return fmt.Errorf("invalid frequency: %s", csvErr)
		}
		refFreqs[strings.ReplaceAll(record[0], `"`, `\"`)] = freq
	}

	wd, err := os.Getwd()
	if err != nil {
		return errors.New("getting working directory " + err.Error())
	}
	wd, err = filepath.EvalSymlinks(wd)
	if err != nil {
		return errors.New("evaluating symlinks " + err.Error())
	}

	done := make(chan struct{})
	defer close(done)
	paths, errc := walkFiles(done, wd)
	c := make(chan result)
	var wg sync.WaitGroup
	const numScanners = 32
	wg.Add(numScanners)
	for i := 0; i < numScanners; i++ {
		go func() {
			entryScanner(done, paths, c)
			wg.Done()
		}()
	}
	go func() {
		wg.Wait()
		close(c)
	}()
	var entryCount float64
	wordMap := make(map[string]int)
	t := time.Now()
	minDate := t
	for r := range c {
		if r.err != nil {
			return r.err
		}
		entryCount++
		for word, count := range r.wordMap {
			wordMap[word] += count
		}
		if minDate.After(r.date) {
			minDate = r.date
		}
	}
	// Check whether the Walk failed.
	if err := <-errc; err != nil {
		return err
	}
	if entryCount > 0 {
		percent := entryCount / math.Floor(t.Sub(minDate).Hours()/24)
		const outFormat = "Jan 2 2006"
		fmt.Fprintf(w, "%.2f%% of days journaled since %v\n", percent*100, minDate.Format(outFormat))
		var wordCount int
		for _, count := range wordMap {
			wordCount += count
		}
		fmt.Fprintf(w, "Total word count: %v\n", wordCount)
		avgCount := float64(wordCount) / entryCount
		fmt.Fprintf(w, "Average word count: %.1f\n", avgCount)
		fmt.Fprint(w, "\n")

		if len(refFreqs) == 0 {
			return nil // no code generation. exit early
		}

		wordStats := make([]*wordStat, len(wordMap))
		i := 0
		for word, count := range wordMap {
			frequency := float64(count) / float64(wordCount)
			var relFrequency float64
			refFrequency := refFreqs[word]
			if frequency > refFrequency {
				if refFrequency > 0 {
					relFrequency = frequency / refFrequency
				}
			} else {
				relFrequency = (refFrequency / frequency) * -1
			}
			wordStats[i] = &wordStat{word: word, occurrences: count, frequency: relFrequency}
			i++
		}

		sort.Slice(wordStats, func(i, j int) bool {
			return wordStats[i].frequency > wordStats[j].frequency
		})

		topUnusualWordCount := 100
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
		fmt.Fprintf(w, "Top %v unusually frequent words:\n", topUnusualWordCount)
		for _, ws := range wordStats[:topUnusualWordCount] {
			fmt.Fprintf(w, "%v\t%.1fX\n", ws.word, ws.frequency)
		}
		w.Flush()
		fmt.Fprint(w, "\n")
		fmt.Fprintf(w, "Top %v unusually infrequent words:\n", topUnusualWordCount)
		for i := 1; i <= topUnusualWordCount; i++ {
			ws := wordStats[len(wordStats)-i]
			fmt.Fprintf(w, "%v\t%.1fX\n", ws.word, ws.frequency)
		}
		w.Flush()
	}
	return nil
}

type result struct {
	wordMap map[string]int
	date    time.Time
	err     error
}

type wordStat struct {
	word        string
	occurrences int
	frequency   float64
}

func walkFiles(
	done <-chan struct{},
	root string,
) (paths chan string, errc chan error) {
	paths = make(chan string)
	errc = make(chan error, 1)
	visited := make(map[string]bool)
	go func() {
		// Close the paths channel after Walk returns.
		defer close(paths)
		// No select needed for this send, since errc is buffered.
		errc <- filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.Mode().IsRegular() || visited[info.Name()] || !IsEntry(path) {
				return nil
			}
			visited[info.Name()] = true
			select {
			case paths <- path:
			case <-done:
				return errors.New("walk canceled")
			}
			return nil
		})
	}()
	return paths, errc
}

func entryScanner(done <-chan struct{}, paths <-chan string, c chan<- result) {
	for path := range paths {
		p := &Entry{Path: path}
		m := make(map[string]int)
		_, err := p.Load()
		if err == nil {
			for _, word := range p.Words() {
				m[strings.ToLower(string(word))]++
			}
		}
		date, _ := p.Date()
		select {
		case c <- result{date: date, wordMap: m, err: err}:
		case <-done:
			return
		}
	}
}
