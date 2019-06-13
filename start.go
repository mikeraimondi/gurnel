package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/mikeraimondi/journalentry/v2"
)

var cmdStart = &command{
	UsageLine: "start",
	ShortHelp: "Begin journal entry for today",
	LongHelp:  "TODO",
	Run:       runStart,
}

func runStart(cmd *command, args []string) error {
	// read in config
	configDir, ok := os.LookupEnv("XDG_CONFIG_HOME")
	if !ok {
		homedir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("getting user home directory: %s", err)
		}
		configDir = filepath.Join(homedir, ".config")
	}
	configData, err := ioutil.ReadFile(filepath.Join(configDir, "gurnel", "config.json"))
	if err != nil {
		// TODO not an error if no config
		return fmt.Errorf("opening config file: %s", err)
	}
	var conf config
	if err := json.Unmarshal(configData, &conf); err != nil {
		return fmt.Errorf("parsing config: %s", err)
	}

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

type beeminderClient struct {
	Token     []byte
	User      string
	c         http.Client
	serverURL string
}

func newBeeminderClient(user string, token []byte) (*beeminderClient, error) {
	if len(user) == 0 {
		return nil, fmt.Errorf("user must not be blank")
	}
	if token != nil && len(token) == 0 {
		return nil, fmt.Errorf("token must not be blank")
	}

	return &beeminderClient{
		Token:     bytes.TrimSpace(token),
		User:      user,
		serverURL: "https://www.beeminder.com",
	}, nil
}

func (client *beeminderClient) postDatapoint(goal string, count int) error {
	if len(goal) == 0 {
		return fmt.Errorf("goal must not be blank")
	}
	if count < 0 {
		return fmt.Errorf("count must be nonnegative")
	}

	postURL, err := url.Parse(client.serverURL)
	if err != nil {
		return fmt.Errorf("internal URL error: %s", err)
	}
	postURL.Path = fmt.Sprintf("api/v1/users/%s/goals/%s/datapoints.json",
		client.User, goal)

	v := url.Values{}
	v.Set("auth_token", string(client.Token))
	v.Set("value", strconv.Itoa(count))
	v.Set("comment", "via Gurnel at "+time.Now().Format("15:04:05 MST"))

	resp, err := client.c.PostForm(postURL.String(), v)
	if err != nil {
		return fmt.Errorf("making request: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		respData, err := ioutil.ReadAll(resp.Body)
		if err != nil || len(respData) == 0 {
			respData = []byte("no further info")
		}
		return fmt.Errorf("server returned %s: %s", resp.Status, respData)
	}
	return nil
}
