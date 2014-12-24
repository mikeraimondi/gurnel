package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type page struct {
	Seconds uint16
	buf     bytes.Buffer
}

func main() {
	t := time.Now()

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
	filename := directory + "/" + t.Format("2006_01_02"+".md")
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
		fmt.Println("File created: " + wd + "/" + file.Name())
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
	var p page
	data, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return
	}
	if err := Unmarshal(data, &p, &p.buf); err != nil {
		fmt.Printf("Error reading YAML frontmatter: %v\n", err)
		return
	}
	p.Seconds += uint16(elapsed.Seconds())
	fmData, err := Marshal(&p, &p.buf)
	if err != nil {
		fmt.Printf("Error writing YAML frontmatter: %v\n", err)
		return
	}
	if _, err := file.WriteAt(fmData, 0); err != nil {
		fmt.Printf("Error writing to file: %v\n", err)
		fmt.Printf("Dump:\n%v\n", fmData)
		return
	}

	// Prompt for commit
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Commit? (y/n)")
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
		fmt.Println("Unrecognized input, exiting")
		return
	}
}
