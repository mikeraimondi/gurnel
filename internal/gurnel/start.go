package gurnel

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/mikeraimondi/journalentry/v2"
)

type startCmd struct{}

func (*startCmd) Name() string       { return "start" }
func (*startCmd) ShortHelp() string  { return "Begin journal entry for today" }
func (*startCmd) LongHelp() string   { return "TODO" }
func (*startCmd) Flag() flag.FlagSet { return flag.FlagSet{} }

func (*startCmd) Run(conf *config, args []string) error {
	// Create or open entry at working directory
	wd, err := os.Getwd()
	if err != nil {
		return errors.New("getting working directory " + err.Error())
	}
	wd, err = filepath.EvalSymlinks(wd)
	if err != nil {
		return errors.New("evaluating symlinks " + err.Error())
	}
	p, err := journalentry.New(wd)
	if err != nil {
		return err
	}

	// Open file for editing
	editor := os.Getenv("GURNEL_EDITOR")
	if len(editor) == 0 {
		editor = os.Getenv("EDITOR")
	}
	editCmd := strings.Split(editor, " ")
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

				token, err := ioutil.ReadFile(conf.BeeminderTokenFile)
				if err != nil {
					return fmt.Errorf("reading token: %s", err)
				}
				client, err := newBeeminderClient(conf.BeeminderUser, token)
				if err != nil {
					return fmt.Errorf("setting up client: %s", err)
				}
				err = client.postDatapoint(conf.BeeminderGoal, wordCount)
				if err != nil {
					return fmt.Errorf("posting to Beeminder: %s", err)
				}

				return nil
			} else if input == "n" {
				fmt.Println("Exiting")
				return nil
			} else {
				fmt.Println("Unrecognized input")
			}
		}
	}

	return nil
}
