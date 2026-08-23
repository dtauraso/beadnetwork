package Dispatch

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"

	NodeBuf "github.com/dtauraso/wirefold/Categories/Node"
)

func nodeIDsInTree(root string) []int {
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

func nodeIDStringsInTree(root string) []string {
	ids := nodeIDsInTree(root)
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		out = append(out, strconv.Itoa(id))
	}
	return out
}

func largestNodeID(root string) int {
	best := 0
	for _, id := range nodeIDsInTree(root) {
		if id > best {
			best = id
		}
	}
	return best
}

func countEdgeFiles(root string) int {
	total := 0
	for _, id := range nodeIDStringsInTree(root) {
		names, err := readDirNames(filepath.Join(root, "nodes", id, "edges"))
		if err != nil {
			continue
		}
		total += len(names)
	}
	return total
}

func newNodeID(root string) string {
	return strconv.Itoa(largestNodeID(root) + 1)
}

func kindForID(id uint8) (string, bool) {
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

func nodesPath(root string) string { return filepath.Join(root, "counts", "nodes.bin") }

func edgesPath(root string) string { return filepath.Join(root, "counts", "edges.bin") }

func writeCounts(root string, nodes, edges int) error {
	if err := NodeBuf.WriteAtomic(nodesPath(root), nodes); err != nil {
		return err
	}
	return NodeBuf.WriteAtomic(edgesPath(root), edges)
}

type Link struct {
	SrcPort    string
	TargetPort string
	Broadcast  bool
}

func linkRefusalFor(src, srcKind string, srcFound bool, kind string) (Link, string, bool) {
	targetPort, hasIn := firstPortOfDir(kind, PortIn)
	if !hasIn {
		return Link{}, fmt.Sprintf("%s takes no input, so nothing can connect to it", kind), false
	}
	if !srcFound {
		return Link{}, fmt.Sprintf("no geometry for %s", src), false
	}
	srcPort, broadcast, hasOut := firstOutputPort(srcKind)
	if !hasOut {
		return Link{}, fmt.Sprintf("%s has no output to connect from", srcKind), false
	}
	return Link{SrcPort: srcPort, TargetPort: targetPort, Broadcast: broadcast}, "", true
}

func firstOutputPort(kind string) (name string, broadcast, ok bool) {
	if name, ok = firstPortOfDir(kind, PortOut); ok {
		return name, false, true
	}
	if name, ok = firstPortOfDir(kind, PortBroadcast); ok {
		return name, true, true
	}
	return "", false, false
}

func firstPortOfDir(kind string, dir PortDir) (string, bool) {
	ports, ok := KindPorts[kind]
	if !ok {
		return "", false
	}
	return FirstPortOfDir(ports, dir)
}
