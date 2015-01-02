package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/mikeraimondi/frontmatter"
)

const (
	wordRegex   = `\w+`
	ratingRegex = `^[1-5]$`
)

type page struct {
	Seconds     uint16
	LowMood     uint8
	HighMood    uint8
	AverageMood uint8
	body        []byte
}

func (p *page) setLowMood(rating uint8) {
	p.LowMood = rating
}

func (p *page) setHighMood(rating uint8) {
	p.HighMood = rating
}

func (p *page) setAvgMood(rating uint8) {
	p.AverageMood = rating
}

func (p *page) prompts() (pr map[string]func(uint8)) {
	pr = make(map[string]func(uint8))
	if p.HighMood == 0 {
		pr["High mood for the day? (1-5) "] = p.setHighMood
	}
	if p.LowMood == 0 {
		pr["Low mood for the day? (1-5) "] = p.setLowMood
	}
	if p.AverageMood == 0 {
		pr["Average mood for the day? (1-5) "] = p.setAvgMood
	}
	return pr
}

func readFile(f *os.File) (p *page, err error) {
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return p, err
	}
	p = &page{}
	if p.body, err = frontmatter.Unmarshal(data, p); err != nil {
		return p, err
	}
	return p, nil
}

func (p *page) writeFile(f *os.File) (err error) {
	fm, err := frontmatter.Marshal(&p)
	if err != nil {
		return err
	}
	if _, err = f.WriteAt(append(fm, p.body...), 0); err != nil {
		fmt.Println("Dump:")
		fmt.Println(string(fm))
		fmt.Println(string(p.body))
	}
	return err
}

func (p *page) words() [][]byte {
	return regexp.MustCompile(wordRegex).FindAll(p.body, -1)
}

func (p *page) promptForMetadata(reader io.Reader, w io.Writer) (err error) {
	r := bufio.NewReader(reader)
	for prompt, setter := range p.prompts() {
		for {
			fmt.Fprint(w, prompt)
			input, err := r.ReadString('\n')
			if err != nil {
				return err
			}
			input = strings.TrimSpace(input)
			regex := regexp.MustCompile(ratingRegex)
			if regex.MatchString(input) {
				rating, err := strconv.ParseUint(input, 10, 8)
				if err != nil {
					return err
				}
				setter(uint8(rating))
				break
			} else {
				fmt.Fprintln(w, "Unrecognized input")
			}
		}
	}
	return err
}
