package test

import (
	"io/ioutil"
	"os"
	"testing"
)

func SetupTestDir(t *testing.T) (dir string, cleanup func()) {
	dir, err := ioutil.TempDir("", "gurnel_test")
	if err != nil {
		t.Fatalf("creating test dir: %s", err)
	}
	if err = os.Chdir(dir); err != nil {
		t.Fatalf("changing to test dir: %s", err)
	}

	cleanup = func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatalf("removing test dir: %s", err)
		}
	}
	return dir, cleanup
}

func WriteFile(t *testing.T, path, contents string) func() {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_APPEND|os.O_SYNC, 0600)
	if err != nil {
		t.Fatalf("opening file: %s", err)
	}

	if _, err = f.WriteString(contents); err != nil {
		t.Fatalf("writing file: %s", err)
	}

	return func() {
		if err := f.Close(); err != nil {
			t.Fatalf("closing file: %s", err)
		}
	}
}
