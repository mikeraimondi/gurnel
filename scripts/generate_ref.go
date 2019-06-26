package main

import (
	"bytes"
	"compress/gzip"
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

func main() {
	if err := generateRef(); err != nil {
		fmt.Println("error: ", err, "; exiting")
		os.Exit(1)
	}
}

func generateRef() error {
	resp, err := http.Get(
		"https://github.com/mikeraimondi/word_frequencies/raw/master/dist/eng-us-10000-1960.csv.gz",
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	zipReader, err := gzip.NewReader(resp.Body)
	if err != nil {
		return err
	}
	defer zipReader.Close()
	csvReader := csv.NewReader(zipReader)
	csvReader.FieldsPerRecord = 2

	buf := &bytes.Buffer{}
	fmt.Fprint(buf, `package gurnel

  func init() {
		refFreqs = map[string]float64{
`)

	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if len(record[0]) == 0 || len(record[1]) == 0 {
			return fmt.Errorf("invalid input")
		}
		fmt.Fprintf(buf, "      \"%v\": %v,\n", strings.Replace(record[0], `"`, `\"`, -1), record[1])
	}

	fmt.Fprint(buf, "  }\n}")
	if err := ioutil.WriteFile("ref_freqs_generated.go", buf.Bytes(), 0644); err != nil {
		return err
	}
	return nil
}
