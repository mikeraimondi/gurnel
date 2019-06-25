package main

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

func (tdp *testDirProvider) getConfigDir() (string, bool) {
	return tdp.configDir, true
}

func (tdp *testDirProvider) getHomeDir() (string, error) {
	return "", nil
}

func TestLoadConfig(t *testing.T) {
	testCases := []struct {
		desc          string
		preLoadConfFn func(string) config
		postLoadConf  config
	}{
		{
			desc: "with an existing valid file",
			preLoadConfFn: func(dir string) config {
				return config{
					dp: &testDirProvider{
						configDir: dir,
					},
				}
			},
			postLoadConf: config{
				BeeminderEnabled: true,
			},
		},
		{
			desc: "with no existing file",
			preLoadConfFn: func(_ string) config {
				return config{
					dp: &testDirProvider{
						configDir: "/not/a/real/directory942310",
					},
				}
			},
			postLoadConf: config{
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
			// TODO delete file
			if err := json.NewEncoder(file).Encode(&tC.postLoadConf); err != nil {
				t.Fatalf("writing to temp file: %s", err)
			}

			c := tC.preLoadConfFn(filepath.Dir(file.Name()))
			if err := c.load(filepath.Base(file.Name())); err != nil {
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
		pre  config
		post config
	}{
		{
			desc: "with no provider",
			pre:  config{},
			post: config{
				dp: &defaultDirProvider{},
			},
		},
		{
			desc: "with a provider",
			pre: config{
				dp: tdp,
			},
			post: config{
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
