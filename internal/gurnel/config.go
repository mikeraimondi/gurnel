package gurnel

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

type dirProvider interface {
	getConfigDir() (string, error)
}

type clock interface {
	Now() time.Time
}

type Config struct {
	BeeminderEnabled   bool
	BeeminderUser      string
	BeeminderTokenFile string
	BeeminderGoal      string
	MinimumWordCount   int
	Editor             string
	dp                 dirProvider
	subcommands        []subcommand
	clock              clock
}

type defaultDirProvider struct{}

func (dp *defaultDirProvider) getConfigDir() (string, error) {
	return os.UserConfigDir()
}

type defaultClock struct{}

func (c *defaultClock) Now() time.Time { return time.Now() }

func (c *Config) Load(path ...string) error {
	c.setupSubcommands()
	c.MinimumWordCount = 750

	if c.clock == nil {
		c.clock = &defaultClock{}
	}

	dir, err := c.getConfigDir()
	if err != nil {
		return fmt.Errorf("getting config directory: %w", err)
	}

	path = append([]string{dir}, path...)
	configData, err := ioutil.ReadFile(filepath.Join(path...))
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("opening file: %w", err)
	}
	return json.Unmarshal(configData, c)
}

func (c *Config) getConfigDir() (string, error) {
	if c.dp == nil {
		c.dp = &defaultDirProvider{}
	}

	dir, err := c.dp.getConfigDir()
	if err != nil {
		return "", err
	}
	return dir, nil
}

func (c *Config) setupSubcommands() {
	if len(c.subcommands) == 0 {
		c.subcommands = []subcommand{
			&startCmd{},
			&statsCmd{},
		}
	}
}
