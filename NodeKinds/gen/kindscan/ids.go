package kindscan

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func AssignKindIDs(kinds []KindEntry, nodesDir string) {
	usedBy := map[uint8]string{}
	maxID := -1
	var unassigned []int

	for i := range kinds {
		raw := strings.TrimSpace(kinds[i].View.KindID)
		if raw == "" {
			unassigned = append(unassigned, i)
			continue
		}
		n, err := strconv.Atoi(raw)
		if err != nil || n < 0 || n > 254 {
			fatalf("kind %q: SPEC.md kindId %q must be an integer in [0,254]", kinds[i].GoKind, raw)
		}
		id := uint8(n)
		if prev, dup := usedBy[id]; dup {
			fatalf("kind %q and kind %q both claim KindId %d in their SPEC.md — ids must be unique and assigned once", prev, kinds[i].GoKind, id)
		}
		usedBy[id] = kinds[i].GoKind
		kinds[i].KindID = id
		if n > maxID {
			maxID = n
		}
	}

	for _, i := range unassigned {
		maxID++
		if maxID > 254 {
			fatalf("kind %q: no free KindId below the KindIDUnknown sentinel (0xFF)", kinds[i].GoKind)
		}
		id := uint8(maxID)
		usedBy[id] = kinds[i].GoKind
		kinds[i].KindID = id
		if err := writeBackKindID(nodesDir, kinds[i].Dir, id); err != nil {
			fatalf("kind %q: auto-assigned KindId %d but failed to write it back into SPEC.md: %v", kinds[i].GoKind, id, err)
		}
		fmt.Fprintf(os.Stderr, "gen-node-defs: auto-assigned KindId %d to new kind %q (written to Node/%s/SPEC.md)\n", id, kinds[i].GoKind, kinds[i].Dir)
	}
}

func writeBackKindID(nodesDir, dir string, id uint8) error {
	specPath := filepath.Join(nodesDir, dir, "SPEC.md")
	data, err := os.ReadFile(specPath)
	if err != nil {
		return err
	}
	lines := strings.Split(string(data), "\n")
	for i, l := range lines {
		trimmed := strings.TrimSpace(l)
		if strings.HasPrefix(trimmed, "| kind |") || strings.HasPrefix(trimmed, "|kind|") {
			row := fmt.Sprintf("| kindId | %d |", id)
			lines = append(lines[:i], append([]string{row}, lines[i:]...)...)
			return os.WriteFile(specPath, []byte(strings.Join(lines, "\n")), 0644)
		}
	}
	return fmt.Errorf("no '| kind |' row found in View table to anchor kindId insertion")
}
