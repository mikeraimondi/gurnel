package main

import (
	"io"
	"io/ioutil"
	"regexp"

	"gopkg.in/yaml.v2"
)

const fmRegex = `(?ms)^\s*---.*---$`

// Unmarshal parses the file containing YAML frontmatter and stores YAML encoded data in the value pointed to by v. The body of the document is written to w.
func Unmarshal(data []byte, v interface{}, w io.Writer) (err error) {
	regex := regexp.MustCompile(fmRegex)
	fmLoc := regex.FindIndex(data)
	if fmLoc == nil {
		_, err = w.Write(data)
		return err
	}
	if err = yaml.Unmarshal(data[fmLoc[0]:fmLoc[1]], v); err != nil {
		return err
	}
	if len(data) <= fmLoc[1]+1 {
		return
	}
	_, err = w.Write(data[fmLoc[1]+1:])
	return err
}

// Marshal returns the YAML-frontmatter containing document from v. It appends the contents of r.
func Marshal(v interface{}, r io.Reader) (data []byte, err error) {
	if data, err = yaml.Marshal(&v); err != nil {
		return data, err
	}
	body, err := ioutil.ReadAll(r)
	if err != nil {
		return data, err
	}
	data = append([]byte("---\n"), data...)
	data = append(data, []byte("---\n")...)
	data = append(data, body...)
	return data, nil
}
