package main

import (
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
	files, err := filepath.Glob("*")
	if err != nil {
		return err
	}
	const dirRegex = `^\d{4}$`
	regex := regexp.MustCompile(dirRegex)
	t := time.Now()
	minYear := t
	var entries uint64
	var wordCount uint64
	wordMap := make(map[string]uint64)
	for _, file := range files {
		if fi, err := os.Lstat(file); err != nil {
			return err
		} else if fi.IsDir() { // Find all directories that look like a year
			if regex.MatchString(file) {
				const dirFormat = "2006"
				yearTime, err := time.Parse(dirFormat, file)
				if err != nil {
					return err
				}
				if yearTime.Year() <= minYear.Year() {
					minYear = yearTime
				}
				const glob = "*.md"
				dirEntries, err := filepath.Glob(file + string(filepath.Separator) + glob)
				if err != nil {
					return err
				}
				entries += uint64(len(dirEntries))
				in := gen(dirEntries...)
				c1 := getWords(in)
				c2 := getWords(in)
				c3 := getWords(in)
				c4 := getWords(in)
				c5 := getWords(in)
				c6 := getWords(in)
				c7 := getWords(in)
				c8 := getWords(in)
				for word := range merge(c1, c2, c3, c4, c5, c6, c7, c8) {
					wordCount++
					wordMap[strings.ToLower(word)]++
				}
			}
		}
	}
	if entries > 0 {
		fEntries := float64(entries)
		percent := fEntries / math.Floor(t.Sub(minYear).Hours()/24)
		const outFormat = "Jan 2 2006"
		fmt.Printf("%.2f%% of days journaled since %v\n", percent*100, minYear.Format(outFormat))
		fmt.Printf("Total word count: %v\n", wordCount)
		avgCount := float64(wordCount) / fEntries
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

func gen(fileNames ...string) <-chan string {
	out := make(chan string)
	go func() {
		for _, fileName := range fileNames {
			out <- fileName
		}
		close(out)
	}()
	return out
}

func getWords(in <-chan string) <-chan string {
	out := make(chan string)
	go func() {
		for fileName := range in {
			f, err := os.Open(fileName)
			if err != nil {
				// return err
			}
			p, err := readFile(f)
			if err != nil {
				// return err
			}
			for _, word := range p.words() {
				out <- string(word)
			}
		}
		close(out)
	}()
	return out
}

func merge(cs ...<-chan string) <-chan string {
	var wg sync.WaitGroup
	out := make(chan string)
	output := func(c <-chan string) {
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
