package kindscan

import (
	"os"
	"path/filepath"
	"strings"
)

// readSpecMDLines reads nodes/<Kind>/SPEC.md and splits it into lines. Shared
// by every SPEC.md parser (parseSpecMD, parsePortsFromSpec, parseDefaultData)
// so the file-read/split step isn't repeated per caller.
func readSpecMDLines(pkgDir string) ([]string, error) {
	data, err := os.ReadFile(filepath.Join(pkgDir, "SPEC.md"))
	if err != nil {
		return nil, err
	}
	return strings.Split(string(data), "\n"), nil
}

// sectionLines returns the lines strictly between a "## heading" line and the
// next "## " heading (or EOF), or nil if heading isn't present.
func sectionLines(lines []string, heading string) []string {
	start := -1
	for i, l := range lines {
		if strings.TrimSpace(l) == "## "+heading {
			start = i
			break
		}
	}
	if start == -1 {
		return nil
	}
	end := len(lines)
	for i := start + 1; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "## ") {
			end = i
			break
		}
	}
	return lines[start+1 : end]
}

// parseMDTable parses a markdown table (pipe-delimited rows, with a
// separator row skipped) into its header cells and data rows.
func parseMDTable(tableLines []string) ([]string, [][]string) {
	var rows []string
	var headers []string
	var result [][]string
	for _, l := range tableLines {
		if !strings.Contains(l, "|") {
			continue
		}
		rows = append(rows, l)
	}
	if len(rows) < 2 {
		return nil, nil
	}
	// First row is headers.
	parts := strings.Split(rows[0], "|")
	for _, p := range parts {
		h := strings.TrimSpace(p)
		if h != "" {
			headers = append(headers, h)
		}
	}
	for _, row := range rows[1:] {
		cells := parseMDRowCells(row)
		// Skip separator rows.
		allSep := true
		for _, c := range cells {
			if !isSep(c) {
				allSep = false
				break
			}
		}
		if allSep {
			continue
		}
		result = append(result, cells)
	}
	return headers, result
}

// parseMDRowCells splits one pipe-delimited markdown table row into trimmed
// cells, dropping the leading/trailing empty cell that a leading/trailing
// "|" produces.
func parseMDRowCells(row string) []string {
	parts := strings.Split(row, "|")
	var cells []string
	for _, p := range parts {
		cells = append(cells, strings.TrimSpace(p))
	}
	if len(cells) > 0 && cells[0] == "" {
		cells = cells[1:]
	}
	if len(cells) > 0 && cells[len(cells)-1] == "" {
		cells = cells[:len(cells)-1]
	}
	return cells
}

// isSep reports whether s is a markdown table separator cell (e.g. "---",
// ":--", "--:"): only '-', ':', and spaces.
func isSep(s string) bool {
	for _, c := range s {
		if c != '-' && c != ':' && c != ' ' {
			return false
		}
	}
	return len(s) > 0
}

func indexOf[T comparable](slice []T, val T) int {
	for i, v := range slice {
		if v == val {
			return i
		}
	}
	return -1
}
