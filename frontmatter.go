package main

import (
	"errors"
	"regexp"

	"gopkg.in/yaml.v2"
)

const fmRegex = `(?ms)^\s*---.*---$`

// Texter is dumb
type Texter interface {
	GetText() []byte

	SetText([]byte)
}

// Unmarshal parses the file containing YAML frontmatter and stores the result in the value pointed to by v.
func Unmarshal(data []byte, v Texter) (err error) {
	regex := regexp.MustCompile(fmRegex)
	if fmLoc := regex.FindIndex(data); fmLoc != nil {
		if err = yaml.Unmarshal(data[fmLoc[0]:fmLoc[1]], v); err != nil {
			return err
		}
		if len(data) > fmLoc[1]+1 {
			v.SetText(data[fmLoc[1]+1:])
		}
	} else {
		return errors.New("No frontmatter found")
	}
	return nil
}

// Marshal returns the YAML-frontmatter containing document from v
func Marshal(v Texter) (data []byte, err error) {
	if data, err = yaml.Marshal(&v); err != nil {
		return data, err
	}
	data = append([]byte("---\n"), data...)
	data = append(data, []byte("---\n")...)
	data = append(data, v.GetText()...)
	return data, nil
}
