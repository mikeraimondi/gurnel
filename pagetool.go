package main

import (
	"fmt"
	"os"
	"time"
)

func main() {
	filename := time.Now().Format("2006_01_02" + ".md")
	if _, err := os.Stat(filename); err == nil {
		fmt.Println("Error: file exists")
		return
	}
	file, err := os.Create(filename)
	if err != nil {
		fmt.Println("Error: " + err.Error())
	}
	defer file.Close()
}
