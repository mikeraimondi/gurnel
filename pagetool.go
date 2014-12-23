package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

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
	filename := directory + "/" + t.Format("2006_01_02"+".md")
	if _, err := os.Stat(filename); err == nil {
		fmt.Println("Error: file exists")
		return
	}

	// Create file
	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Error creating file: %v\n", err)
		return
	}
	defer file.Close()
	wd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting working directory: %v\n", err)
		return
	}
	fmt.Println("File created: " + wd + "/" + file.Name())

	// Open file for editing
	editCmd := strings.Split(os.Getenv("EDITOR"), " ")
	editCmd = append(editCmd, file.Name())
	startTime := time.Now()
	if err := exec.Command(editCmd[0], editCmd[1:]...).Run(); err != nil {
		fmt.Printf("Error opening editor: %v\n", err)
		return
	}

	// Abort and remove file if untouched
	fi, err := file.Stat()
	if err != nil {
		fmt.Printf("Error getting file info: %v\n", err)
		return
	}
	if fi.Size() == 0 {
		fmt.Println("Aborting due to empty file")
		os.Remove(file.Name())
		return
	}

	// Add metadata to front matter
	elapsed := time.Since(startTime)
	buf := bytes.NewBuffer(nil)
	fmt.Fprintf(buf, "---\nseconds: %.0f\n---\n", elapsed.Seconds())
	if _, err := io.Copy(buf, file); err != nil {
		fmt.Printf("Error writing front matter: %v\n", err)
		return
	}
	if _, err := file.WriteAt(buf.Bytes(), 0); err != nil {
		fmt.Printf("Error writing to file: %v\n", err)
		fmt.Printf("Dump:\n%v\n", buf)
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
