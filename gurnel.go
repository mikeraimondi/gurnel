package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const dirRegex = `^\d{4}$`

func main() {
	t := time.Now()

	if len(os.Args) > 1 && os.Args[1] == "percent" {
		files, _ := filepath.Glob("*")
		regex := regexp.MustCompile(dirRegex)
		minYear := t
		entries := 0
		for _, file := range files {
			if fi, err := os.Lstat(file); err != nil {
				fmt.Printf("Error stat'ing file: %v\n", err)
			} else if fi.IsDir() {
				if regex.MatchString(file) {
					yearTime, _ := time.Parse("2006", file)
					if yearTime.Year() <= minYear.Year() {
						minYear = yearTime
					}
					dirEntries, _ := filepath.Glob(file + string(filepath.Separator) + "*.md")
					entries += len(dirEntries)
				}
			}
		}
		if entries > 0 {
			percent := float64(entries) / math.Ceil(t.Sub(minYear).Hours()/24)
			fmt.Println(percent)
		}
		return
	}

	// Create directory if it doesn't exist
	directory := strconv.Itoa(t.Year())
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		if err := os.Mkdir(directory, 0755); err != nil {
			fmt.Printf("Error creating directory: %v\n", err)
			return
		}
	}

	// Test for presence of file
	var file *os.File
	defer file.Close()
	os.Chdir(directory)
	filename := t.Format("2006_01_02" + ".md")
	if _, err := os.Stat(filename); err != nil {
		// Create file
		if file, err = os.Create(filename); err != nil {
			fmt.Printf("Error creating file: %v\n", err)
			return
		}
		wd, err := os.Getwd()
		if err != nil {
			fmt.Printf("Error getting working directory: %v\n", err)
			return
		}
		fmt.Println("File created: " + wd + string(filepath.Separator) + file.Name())
	} else {
		// Open file
		var perm os.FileMode = 0666
		if file, err = os.OpenFile(filename, os.O_RDWR, perm); err != nil {
			fmt.Printf("Error opening file: %v\n", err)
			return
		}
	}

	// Store modification time
	fi, err := file.Stat()
	if err != nil {
		fmt.Printf("Error getting file info: %v\n", err)
		return
	}
	modTime := fi.ModTime()

	// Open file for editing
	editCmd := strings.Split(os.Getenv("EDITOR"), " ")
	editCmd = append(editCmd, file.Name())
	startTime := time.Now()
	if err := exec.Command(editCmd[0], editCmd[1:]...).Run(); err != nil {
		fmt.Printf("Error opening editor: %v\n", err)
		return
	}
	elapsed := time.Since(startTime)

	// Abort if file is untouched
	if fi, err = file.Stat(); err != nil {
		fmt.Printf("Error getting file info: %v\n", err)
		return
	}
	if fi.ModTime() == modTime {
		fmt.Println("Aborting due to unchanged file")
		return
	}

	// Parse & process frontmatter
	p, err := readFile(file)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return
	}
	p.Seconds += uint16(elapsed.Seconds())
	if err := p.writeFile(file); err != nil {
		fmt.Printf("Error writing to file: %v\n", err)
		return
	}

	// Prompt for commit
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Commit? (y/n) ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input == "y" {
			// Commit the changes
			err = exec.Command("git", "add", file.Name()).Run()
			if err != nil {
				fmt.Printf("Error adding file to version control: %v\n", err)
				return
			}
			err = exec.Command("git", "commit", "-m", "Done").Run()
			if err != nil {
				fmt.Printf("Error committing file: %v\n", err)
				return
			}
			fmt.Println("Committed")
			return
		} else if input == "n" {
			fmt.Println("Exiting")
			return
		} else {
			fmt.Println("Unrecognized input")
		}
	}
}
