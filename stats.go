package main

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
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
	entries := 0
	wordCount := 0
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
				entries += len(dirEntries)
				for _, fileName := range dirEntries { // Get wordcount of each entry
					f, err := os.Open(fileName)
					if err != nil {
						return err
					}
					p, err := readFile(f)
					if err != nil {
						return err
					}
					wordCount += len(p.words())
				}
			}
		}
	}
	if entries > 0 {
		fEntries := float64(entries)
		percent := fEntries / math.Floor(t.Sub(minYear).Hours()/24)
		const outFormat = "Jan 2 2006"
		fmt.Printf("%.2f%% of days journaled since %v\n", percent*100, minYear.Format(outFormat))
		avgCount := float64(wordCount) / fEntries
		fmt.Printf("Average word count: %.1f\n", avgCount)
	}
	return
}
