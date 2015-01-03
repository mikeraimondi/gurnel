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
		var wordStats []wordStat
		for word, count := range wordMap {
			wordStats = append(wordStats, wordStat{word: word, occurrences: count})
		}
		sort.Sort(byOccurences(wordStats))
		commonWords := []string{"i", "to", "the", "a", "and", "of", "is", "that", "in", "for", "it", "be", "on", "at", "this", "with", "but", "not"}
		i := 0
		for _, ws := range wordStats {
			common := false
			for _, cw := range commonWords {
				if ws.word == cw {
					common = true
					break
				}
			}
			if !common && len(ws.word) > 2 {
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

type byOccurences []wordStat

func (o byOccurences) Len() int           { return len(o) }
func (o byOccurences) Swap(i, j int)      { o[i], o[j] = o[j], o[i] }
func (o byOccurences) Less(i, j int) bool { return o[i].occurrences > o[j].occurrences }
