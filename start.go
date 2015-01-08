package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/mikeraimondi/journalentry"
)

func startCmd() gurnelCmd {
	return gurnelCmd{
		f:             start,
		condensedHelp: "Begin journal entry for today",
		fullHelp:      "TODO",
	}
}

func start(args []string) (err error) {
	// Create or open entry at working directory
	wd, err := os.Getwd()
	if err != nil {
		return errors.New("getting working directory " + err.Error())
	}
	p, err := journalentry.New(wd)
	if err != nil {
		return err
	}

	// Open file for editing
	editCmd := strings.Split(os.Getenv("EDITOR"), " ")
	editCmd = append(editCmd, p.Path)
	startTime := time.Now()
	if err := exec.Command(editCmd[0], editCmd[1:]...).Run(); err != nil {
		return errors.New("opening editor " + err.Error())
	}
	elapsed := time.Since(startTime)

	// Abort if file is untouched
	if modified, err := p.Load(); err != nil {
		return errors.New("loading file " + err.Error())
	} else if !modified {
		fmt.Println("Aborting due to unchanged file")
		return nil
	}

	// Check word count before proceeding to metadata collection
	wordCount := len(p.Words())
	fmt.Printf("%v words in entry\n", wordCount)
	if wordCount < minWordCount {
		fmt.Printf("Minimum word count is %v. Insufficient word count to commit\n", minWordCount)
	} else {
		fmt.Printf("---begin entry preview---\n%v\n--end entry preview---\n", string(p.Body))

		// Collect & set metadata
		if err := p.PromptForMetadata(os.Stdin, os.Stdout); err != nil {
			return errors.New("collecting metadata " + err.Error())
		}
	}
	p.Seconds += uint16(elapsed.Seconds())
	if err := p.Save(); err != nil {
		return errors.New("saving file " + err.Error())
	}

	if wordCount > minWordCount {
		// Prompt for commit
		reader := bufio.NewReader(os.Stdin)
		for {
			fmt.Print("Commit? (y/n) ")
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(input)
			if input == "y" {
				// Commit the changes
				err = exec.Command("git", "add", p.Path).Run()
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
	return
}
