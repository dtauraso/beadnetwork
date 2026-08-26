package main

import "strings"

func declaresNoPorts(pkgDir string) bool {
	lines, err := readSpecMDLines(pkgDir)
	if err != nil {
		return false
	}
	for _, line := range sectionLines(lines, "Ports") {
		text := strings.TrimSpace(line)
		if text == "" {
			continue
		}
		return strings.HasPrefix(text, "None")
	}
	return false
}
