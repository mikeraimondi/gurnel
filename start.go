package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func startCmd() gurnelCmd {
	return gurnelCmd{
		f:             start,
		condensedHelp: "Begin journal entry for today",
		fullHelp:      "TODO",
	}
}

func start(args []string) (err error) {
	// Create directory if it doesn't exist
	t := time.Now()
	directory := strconv.Itoa(t.Year())
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		if err := os.Mkdir(directory, 0755); err != nil {
			return errors.New("creating directory " + err.Error())
		}
	}

	// Test for presence of file
	var file *os.File
	defer file.Close()
	os.Chdir(directory)
	const fFormat = "2006_01_02" + ".md"
	filename := t.Format(fFormat)
	if _, err := os.Stat(filename); err != nil {
		// Create file
		if file, err = os.Create(filename); err != nil {
			return errors.New("creating file " + err.Error())
		}
		wd, err := os.Getwd()
		if err != nil {
			return errors.New("getting working directory " + err.Error())
		}
		fmt.Println("File created: " + wd + string(filepath.Separator) + file.Name())
	} else {
		// Open file
		var perm os.FileMode = 0666
		if file, err = os.OpenFile(filename, os.O_RDWR, perm); err != nil {
			return errors.New("opening file " + err.Error())
		}
	}

	// Store modification time
	fi, err := file.Stat()
	if err != nil {
		return errors.New("getting file info " + err.Error())
	}
	modTime := fi.ModTime()

	// Open file for editing
	editCmd := strings.Split(os.Getenv("EDITOR"), " ")
	editCmd = append(editCmd, file.Name())
	startTime := time.Now()
	if err := exec.Command(editCmd[0], editCmd[1:]...).Run(); err != nil {
		return errors.New("opening editor " + err.Error())
	}
	elapsed := time.Since(startTime)

	// Abort if file is untouched
	if fi, err = file.Stat(); err != nil {
		return errors.New("getting file info " + err.Error())
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
		return errors.New("writing to file " + err.Error())
	}

	// TODO prompt for metadata

	// Prompt for commit
	wordCount := len(p.words())
	fmt.Printf("%v words in entry. ", wordCount)
	if wordCount < minWordCount {
		fmt.Printf("Minimum word count is %v. Insufficient word count to commit\n", minWordCount)
		return
	}
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Commit? (y/n) ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input == "y" {
			// Commit the changes
			err = exec.Command("git", "add", file.Name()).Run()
			if err != nil {
				return errors.New("adding file to version control " + err.Error())
			}
			err = exec.Command("git", "commit", "-m", "Done").Run()
			if err != nil {
				return errors.New("committing file " + err.Error())
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
