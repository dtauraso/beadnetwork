package structuraledit

import (
	"os"
	"path/filepath"
	"sort"
	"strconv"

	NodeBuf "github.com/dtauraso/wirefold/Categories/Node"
)

func NodeIDsInTree(root string) []int {
	names, err := readDirNames(filepath.Join(root, "nodes"))
	if err != nil {
		return nil
	}
	out := make([]int, 0, len(names))
	for _, n := range names {
		if v, err := strconv.Atoi(n); err == nil {
			out = append(out, v)
		}
	}
	sort.Ints(out)
	return out
}

func NodeIDStringsInTree(root string) []string {
	ids := NodeIDsInTree(root)
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		out = append(out, strconv.Itoa(id))
	}
	return out
}

func LargestNodeID(root string) int {
	best := 0
	for _, id := range NodeIDsInTree(root) {
		if id > best {
			best = id
		}
	}
	return best
}

func CountEdgeFiles(root string) int {
	total := 0
	for _, id := range NodeIDStringsInTree(root) {
		names, err := readDirNames(filepath.Join(root, "nodes", id, "edges"))
		if err != nil {
			continue
		}
		total += len(names)
	}
	return total
}

func NewNodeID(root string) string {
	return strconv.Itoa(LargestNodeID(root) + 1)
}

func KindForID(id uint8) (string, bool) {
	for _, k := range NodeBuf.KnownKinds() {
		if NodeBuf.NodeKindID(k) == id {
			return k, true
		}
	}
	return "", false
}

func readDirNames(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names, nil
}
