package main

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

func main() {
	t := time.Now()

	// Create directory if it doesn't exist
	directory := strconv.Itoa(t.Year())
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		os.Mkdir(directory, 0755)
	}

	// Test for presence of file
	filename := directory + "/" + t.Format("2006_01_02"+".md")
	if _, err := os.Stat(filename); err == nil {
		fmt.Println("Error: file exists")
		return
	}

	// Create file
	file, err := os.Create(filename)
	if err != nil {
		fmt.Println("Error: " + err.Error())
	}
	defer file.Close()

	// Output filename
	fmt.Println(file.Name())
}
