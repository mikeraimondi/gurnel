package main

import (
	"fmt"
	"os"

	"github.com/mikeraimondi/gurnel/internal/gurnel"
)

func main() {
	var conf gurnel.Config
	if err := conf.Load("gurnel", "gurnel.json"); err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %s\n", err)
		os.Exit(2)
	}

	if err := gurnel.Do(os.Stdin, os.Stdout, &conf); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(2)
	}
}
