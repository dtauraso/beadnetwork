package kindscan

import "strings"

func parseDefaultData(pkgDir string) string {
	lines, err := readSpecMDLines(pkgDir)
	if err != nil {
		return ""
	}
	inSection := false
	inFence := false
	var jsonLines []string
	for _, l := range lines {
		if strings.TrimSpace(l) == "## Default data" {
			inSection = true
			continue
		}
		if inSection && strings.HasPrefix(l, "## ") {
			break
		}
		if inSection && !inFence && strings.TrimSpace(l) == "```json" {
			inFence = true
			continue
		}
		if inSection && inFence {
			if strings.TrimSpace(l) == "```" {
				break
			}
			jsonLines = append(jsonLines, l)
		}
	}
	return strings.TrimSpace(strings.Join(jsonLines, "\n"))
}
