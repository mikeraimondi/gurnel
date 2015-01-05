package main

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

const commonWords = "i to the a and of is that in for it be on at this with but not"

var commonWordsArray []string

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
	t := time.Now()
	minDate := t
	var wordCount uint64
	wordMap := make(map[string]uint64)
	entries, err := filepath.Glob(entryGlob)
	if err != nil {
		return err
	}
	in := gen(entries...)
	c1 := openPage(in)
	c2 := openPage(in)
	c3 := openPage(in)
	c4 := openPage(in)
	c5 := openPage(in)
	c6 := openPage(in)
	c7 := openPage(in)
	c8 := openPage(in)
	for p := range merge(c1, c2, c3, c4, c5, c6, c7, c8) {
		// TODO make concurrent
		if pDate, err := p.date(); err != nil {
			return err
		} else if minDate.After(pDate) {
			minDate = pDate
		}
		for _, w := range p.words() {
			wordCount++
			wordMap[strings.ToLower(string(w))]++
		}
	}
	if entryCount := float64(len(entries)); entryCount > 0 {
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

func gen(fileNames ...string) <-chan *page {
	out := make(chan *page)
	go func() {
		for _, fileName := range fileNames {
			out <- &page{file: fileName}
		}
		close(out)
	}()
	return out
}

func openPage(in <-chan *page) <-chan *page {
	out := make(chan *page)
	go func() {
		for p := range in {
			f, err := os.Open(p.file)
			if err != nil {
				return // err
			}
			p, err := readFile(f)
			if err != nil {
				return // err
			}
			out <- p
		}
		close(out)
	}()
	return out
}

func merge(cs ...<-chan *page) <-chan *page {
	var wg sync.WaitGroup
	out := make(chan *page)
	output := func(c <-chan *page) {
		for n := range c {
			out <- n
		}
		wg.Done()
	}
	wg.Add(len(cs))
	for _, c := range cs {
		go output(c)
	}
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

type wordStat struct {
	word        string
	occurrences uint64
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
