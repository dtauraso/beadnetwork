package panelgen

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

var flagLine = regexp.MustCompile(`^"([A-Za-z0-9]+)",$`)

func ParsePanelFlags(messagesPath string) ([]string, error) {
	data, err := os.ReadFile(messagesPath)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(data), "\n")

	start, end := -1, -1
	for i, l := range lines {
		switch strings.TrimSpace(l) {
		case "// PANEL_FLAGS_START":
			if start == -1 {
				start = i
			}
		case "// PANEL_FLAGS_END":
			if start != -1 && end == -1 {
				end = i
			}
		}
	}
	if start == -1 || end == -1 {
		return nil, fmt.Errorf("no PANEL_FLAGS_START/END block in %s", messagesPath)
	}

	var flags []string
	for _, l := range lines[start:end] {
		m := flagLine.FindStringSubmatch(strings.TrimSpace(l))
		if m != nil {
			flags = append(flags, m[1])
		}
	}
	if len(flags) == 0 {
		return nil, fmt.Errorf("PANEL_FLAGS block in %s names no flags", messagesPath)
	}
	return flags, nil
}
