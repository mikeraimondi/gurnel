package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"

	"github.com/mikeraimondi/frontmatter"
)

const wordRegex = `\w+`

type page struct {
	Seconds uint16
	body    []byte
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

func (p *page) writeFile(f *os.File) error {
	fm, err := frontmatter.Marshal(&p)
	if err != nil {
		return err
	}
	if _, err := f.WriteAt(append(fm, p.body...), 0); err != nil {
		fmt.Println("Dump:")
		fmt.Println(string(fm))
		fmt.Println(string(p.body))
		return err
	}
	return nil
}

func (p *page) words() [][]byte {
	return regexp.MustCompile(wordRegex).FindAll(p.body, -1)
}
