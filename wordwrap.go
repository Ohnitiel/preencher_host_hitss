package main

import "strings"

func Wordwrap(s string, maxWidth int) string {
	var lines []string

	for _, word := range strings.Split(s, " ") {
		if len(lines) == 0 {
			lines = append(lines, word)
		} else {
			if len(lines[len(lines)-1])+len(word) > maxWidth {
				lines = append(lines, word)
			} else {
				lines[len(lines)-1] += " " + word
			}
		}
	}

	return strings.Join(lines, "\n")
}
