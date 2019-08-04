package gurnel

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type startCmd struct{}

func (*startCmd) Name() string       { return "start" }
func (*startCmd) ShortHelp() string  { return "Begin journal entry for today" }
func (*startCmd) Flag() flag.FlagSet { return flag.FlagSet{} }

func (*startCmd) LongHelp() string {
	return "If you don't like the editor this uses, set $EDITOR to something else."
}

func (*startCmd) Run(r io.Reader, w io.Writer, args []string, conf *config) error {
	// Create or open entry at working directory
	wd, err := os.Getwd()
	if err != nil {
		return errors.New("getting working directory " + err.Error())
	}
	wd, err = filepath.EvalSymlinks(wd)
	if err != nil {
		return errors.New("evaluating symlinks " + err.Error())
	}
	p, err := NewEntry(wd)
	if err != nil {
		return err
	}

	// Open file for editing
	editor := conf.Editor
	if editor == "" {
		editor = os.Getenv("EDITOR")
	}
	editCmd := strings.Split(editor, " ")
	editCmd = append(editCmd, p.Path)
	startTime := time.Now()
	cmd := exec.Command(editCmd[0], editCmd[1:]...)
	cmd.Stdin = r
	cmd.Stdout = w
	cmd.Stderr = w
	if err = cmd.Run(); err != nil {
		return errors.New("opening editor " + err.Error())
	}
	elapsed := time.Since(startTime)

	// Abort if file is untouched
	if modified, modErr := p.Load(); modErr != nil {
		return errors.New("loading file " + modErr.Error())
	} else if !modified {
		fmt.Fprintln(w, "Aborting due to unchanged file")
		return nil
	}

	// Check word count before proceeding to metadata collection
	wordCount := len(p.Words())
	fmt.Fprintf(w, "%v words in entry\n", wordCount)
	if wordCount < conf.MinimumWordCount {
		fmt.Fprintf(w, "Minimum word count is %v. Insufficient word count to commit\n", conf.MinimumWordCount)
	} else {
		fmt.Fprintf(w, "---begin entry preview---\n%v\n--end entry preview---\n", string(p.Body))

		// Collect & set metadata
		if promptErr := p.PromptForMetadata(r, w); promptErr != nil {
			return errors.New("collecting metadata " + promptErr.Error())
		}
	}
	p.Seconds += uint16(elapsed.Seconds())
	if saveErr := p.Save(); saveErr != nil {
		return errors.New("saving file " + saveErr.Error())
	}

	if wordCount >= conf.MinimumWordCount {
		// Prompt for commit
		fmt.Fprint(w, "Commit? (y/n) ")
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			input := scanner.Text()
			input = strings.TrimSpace(input)
			switch input {
			case "y":
				// Commit the changes
				err = exec.Command("git", "add", p.Path).Run()
				if err != nil {
					return errors.New("adding file to version control " + err.Error())
				}
				err = exec.Command("git", "commit", "-m", "Done").Run()
				if err != nil {
					return errors.New("committing file " + err.Error())
				}
				fmt.Fprintln(w, "Committed")

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
			case "n":
				fmt.Fprintln(w, "Exiting")
				return nil
			default:
				fmt.Fprintln(w, "Unrecognized input")
				fmt.Fprint(w, "Commit? (y/n) ")
			}
		}
		if scanner.Err() != nil {
			return scanner.Err()
		}
	}
	return nil
}
