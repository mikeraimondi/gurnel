package main

import (
	"bufio"
	"fmt"
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
		os.Mkdir(directory, 0755)
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
	file.Close()
	fmt.Printf("File created: %v\n", file.Name())

	// Open file for editing
	editCmd := strings.Split(os.Getenv("EDITOR"), " ")
	editCmd = append(editCmd, file.Name())
	err = exec.Command(editCmd[0], editCmd[1:]...).Run()
	if err != nil {
		fmt.Printf("Error opening editor: %v\n", err)
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
