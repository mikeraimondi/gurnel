package gurnel

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

type dirProvider interface {
	getConfigDir() (string, bool)
	getHomeDir() (string, error)
}

type config struct {
	BeeminderEnabled   bool
	BeeminderUser      string
	BeeminderTokenFile string
	BeeminderGoal      string
	MinimumWordCount   int
	Editor             string
	dp                 dirProvider
	subcommands        []subcommand
}

type defaultDirProvider struct{}

func (dp *defaultDirProvider) getConfigDir() (string, bool) {
	return os.LookupEnv("XDG_CONFIG_HOME")
}

func (dp *defaultDirProvider) getHomeDir() (string, error) {
	return os.UserHomeDir()
}

func (c *config) load(path ...string) error {
	c.setupSubcommands()
	c.MinimumWordCount = 750

	dir, err := c.getConfigDir()
	if err != nil {
		return fmt.Errorf("getting config directory: %s", err)
	}

	path = append([]string{dir}, path...)
	configData, err := ioutil.ReadFile(filepath.Join(path...))
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("opening file: %s", err)
	}
	return json.Unmarshal(configData, c)
}

func (c *config) getConfigDir() (string, error) {
	if c.dp == nil {
		c.dp = &defaultDirProvider{}
	}

	dir, ok := c.dp.getConfigDir()
	if ok {
		return dir, nil
	}

	homeDir, err := c.dp.getHomeDir()
	if err != nil {
		return "", fmt.Errorf("getting home directory: %s", err)
	}
	return filepath.Join(homeDir, ".config"), nil
}

func (c *config) setupSubcommands() {
	if len(c.subcommands) == 0 {
		c.subcommands = []subcommand{
			&startCmd{},
			&statsCmd{},
		}
	}
}
