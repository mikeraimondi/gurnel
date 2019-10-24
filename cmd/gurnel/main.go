package main

import (
	"fmt"
	"os"
	"log"

	"github.com/mikeraimondi/gurnel/internal/gurnel"
)

func main() {
	if err := gurnel.Do(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(2)
	}
}
