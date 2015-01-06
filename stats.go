package main

import (
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

const commonWords = "i to the a and of is that in for it be on at this with but not"

var commonWordsArray []string

// TODO words should be communicated differently
type result struct {
	page  *page
	words [][]byte
	err   error
}

type wordStat struct {
	word        string
	occurrences uint64
}

func getCommonWords() []string {
	if commonWordsArray == nil {
		commonWordsArray = strings.Split(commonWords, " ")
	}
	return commonWordsArray
}

func statsCmd() gurnelCmd {
	return gurnelCmd{
		f:             stats,
		condensedHelp: "View journal statistics",
		fullHelp:      "TODO",
	}
}

func stats(args []string) (err error) {
	done := make(chan struct{})
	defer close(done)

	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	paths, errc := walkFiles(done, wd)
	c := make(chan result)
	var wg sync.WaitGroup
	const numDigesters = 20
	wg.Add(numDigesters)
	for i := 0; i < numDigesters; i++ {
		go func() {
			digester(done, paths, c)
			wg.Done()
		}()
	}
	go func() {
		wg.Wait()
		close(c)
	}()
	var entryCount float64
	var wordCount uint64
	wordMap := make(map[string]uint64)
	t := time.Now()
	minDate := t
	for r := range c {
		if r.err != nil {
			return r.err
		}
		entryCount++
		for _, w := range r.words {
			wordCount++
			wordMap[strings.ToLower(string(w))]++
		}
		if pDate, err := r.page.date(); err != nil {
			return err
		} else if minDate.After(pDate) {
			minDate = pDate
		}
	}
	// Check whether the Walk failed.
	if err := <-errc; err != nil {
		return err
	}
	if entryCount > 0 {
		percent := entryCount / math.Floor(t.Sub(minDate).Hours()/24)
		const outFormat = "Jan 2 2006"
		fmt.Printf("%.2f%% of days journaled since %v\n", percent*100, minDate.Format(outFormat))
		fmt.Printf("Total word count: %v\n", wordCount)
		avgCount := float64(wordCount) / entryCount
		fmt.Printf("Average word count: %.1f\n", avgCount)

		topWordCount := 10
		fmt.Printf("Top %v words by frequency:\n", topWordCount)
		wordStats := make([]wordStat, len(wordMap))
		i := 0
		for word, count := range wordMap {
			wordStats[i] = wordStat{word: word, occurrences: count}
			i++
		}
		sort.Sort(descOccurences(wordStats))
		i = 0
		for _, ws := range wordStats {
			if len(ws.word) > 2 && !ws.isCommon() {
				i++
				fmt.Println(ws.word)
			}
			if i >= topWordCount {
				break
			}
		}
	}
	return
}

func walkFiles(done <-chan struct{}, root string) (<-chan string, <-chan error) {
	paths := make(chan string)
	errc := make(chan error, 1)
	visited := make(map[string]bool)
	go func() {
		// Close the paths channel after Walk returns.
		defer close(paths)
		// No select needed for this send, since errc is buffered.
		errc <- filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.Mode().IsRegular() || visited[info.Name()] || !regexp.MustCompile(entryRegex).MatchString(path) {
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

func digester(done <-chan struct{}, paths <-chan string, c chan<- result) {
	for path := range paths {
		p, err := fromFile(path)
		select {
		case c <- result{p, p.words(), err}:
		case <-done:
			return
		}
	}
}

func (ws *wordStat) isCommon() bool {
	for _, cw := range getCommonWords() {
		if ws.word == cw {
			return true
		}
	}
	return false
}

type descOccurences []wordStat

func (o descOccurences) Len() int           { return len(o) }
func (o descOccurences) Swap(i, j int)      { o[i], o[j] = o[j], o[i] }
func (o descOccurences) Less(i, j int) bool { return o[i].occurrences > o[j].occurrences }
