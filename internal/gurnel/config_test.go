package gurnel

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"testing"
)

type testDirProvider struct {
	configDir string
}

func (tdp *testDirProvider) getConfigDir() (string, error) {
	return tdp.configDir, nil
}

func TestLoadConfig(t *testing.T) {
	testCases := []struct {
		desc          string
		preLoadConfFn func(string) Config
		postLoadConf  Config
	}{
		{
			desc: "with an existing valid file",
			preLoadConfFn: func(dir string) Config {
				return Config{
					dp: &testDirProvider{
						configDir: dir,
					},
				}
			},
			postLoadConf: Config{
				BeeminderEnabled: true,
			},
		},
		{
			desc: "with no existing file",
			preLoadConfFn: func(_ string) Config {
				return Config{
					dp: &testDirProvider{
						configDir: "/not/a/real/directory942310",
					},
				}
			},
			postLoadConf: Config{
				BeeminderEnabled: false,
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			file, err := ioutil.TempFile("", "testconf")
			if err != nil {
				t.Fatalf("creating temp file: %s", err)
			}
			defer file.Close()
			if err := json.NewEncoder(file).Encode(&tC.postLoadConf); err != nil {
				t.Fatalf("writing to temp file: %s", err)
			}

			c := tC.preLoadConfFn(filepath.Dir(file.Name()))
			if err := c.Load(filepath.Base(file.Name())); err != nil {
				t.Fatalf("expected no error loading config. got %s", err)
			}

			if tC.postLoadConf.BeeminderEnabled != c.BeeminderEnabled {
				t.Fatalf("wrong value for BeeminderEnabled. expected %t. got %t",
					tC.postLoadConf.BeeminderEnabled, c.BeeminderEnabled)
			}
		})
	}
}

func TestGetConfigDir(t *testing.T) {
	tdp := &testDirProvider{
		configDir: filepath.Dir("/tmp"),
	}
	testCases := []struct {
		desc string
		pre  Config
		post Config
	}{
		{
			desc: "with no provider",
			pre:  Config{},
			post: Config{
				dp: &defaultDirProvider{},
			},
		},
		{
			desc: "with a provider",
			pre: Config{
				dp: tdp,
			},
			post: Config{
				dp: tdp,
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			tC.pre.getConfigDir()
			if !reflect.DeepEqual(tC.pre, tC.post) {
				t.Fatalf("wrong config after load.\nexpected: %+v\ngot: %+v",
					tC.post, tC.pre)
			}
		})
	}
}
