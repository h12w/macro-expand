// Copyright 2014, Hǎiliàng Wáng. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
)

var options struct {
	InputFile  string
	MacroFile  string
	OutputFile string
}

func init() {
	flag.StringVar(&options.InputFile, "i", "", "Input file to be proccessed.")
	flag.StringVar(&options.OutputFile, "o", "", "Output file to be saved.")
	flag.StringVar(&options.MacroFile, "m", "", "Macro file.")
}

func main() {
	flag.Parse()
	macros := readMacros()

	f, err := os.Open(options.InputFile)
	c(err)
	buf, err := ioutil.ReadAll(f)
	c(err)

	result := Expand(string(buf), func(key string) string {
		if val, ok := macros[key]; ok {
			return val
		}
		return "${" + key + "}"
	})

	err = ioutil.WriteFile(options.OutputFile, []byte(result), 0664)
	c(err)
}

func readMacros() map[string]string {
	macros := make(map[string]string)
	f, err := os.Open(options.MacroFile)
	c(err)
	err = json.NewDecoder(f).Decode(&macros)
	c(err)
	return macros
}

func c(err error) {
	if err != nil {
		panic(err)
	}
}

func p(v interface{}) {
	b, err := json.MarshalIndent(v, "", "  ")
	c(err)
	fmt.Println(string(b))
}

// Expand replaces ${var} or $var in the string based on the mapping function.
// For example, os.ExpandEnv(s) is equivalent to os.Expand(s, os.Getenv).
func Expand(s string, mapping func(string) string) string {
	buf := make([]byte, 0, 2*len(s))
	// ${} is all ASCII, so bytes are fine for this operation.
	i := 0
	for j := 0; j < len(s); j++ {
		if s[j] == '$' && j+2 < len(s) && s[j+1] == '{' {
			buf = append(buf, s[i:j]...)
			name, w := getShellName(s[j+1:])
			buf = append(buf, mapping(name)...)
			j += w
			i = j + 1
		}
	}
	return string(buf) + s[i:]
}

// isSpellSpecialVar reports whether the character identifies a special
// shell variable such as $*.
func isShellSpecialVar(c uint8) bool {
	switch c {
	case '*', '#', '$', '@', '!', '?', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return true
	}
	return false
}

// isAlphaNum reports whether the byte is an ASCII letter, number, or underscore
func isAlphaNum(c uint8) bool {
	return c == '_' || '0' <= c && c <= '9' || 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z'
}

// getName returns the name that begins the string and the number of bytes
// consumed to extract it.  If the name is enclosed in {}, it's part of a ${}
// expansion and two more bytes are needed than the length of the name.
func getShellName(s string) (string, int) {
	switch {
	case s[0] == '{':
		if len(s) > 2 && isShellSpecialVar(s[1]) && s[2] == '}' {
			return s[1:2], 3
		}
		// Scan to closing brace
		for i := 1; i < len(s); i++ {
			if s[i] == '}' {
				return s[1:i], i + 1
			}
		}
		return "", 1 // Bad syntax; just eat the brace.
	case isShellSpecialVar(s[0]):
		return s[0:1], 1
	}
	// Scan alphanumerics.
	var i int
	for i = 0; i < len(s) && isAlphaNum(s[i]); i++ {
	}
	return s[:i], i
}
