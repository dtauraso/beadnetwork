package Topology

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	NodeBuf "github.com/dtauraso/beadnetwork/Categories/Node"
)

const (
	DirNodeData = "data"
	DirInit     = "init"
	DirState    = "state"
	DirSendRule = "send-rules"
	FileLabel   = "label.bin"
	FileRepeat  = "repeat.bin"
)

func trimLeafExt(name string) string { return strings.TrimSuffix(name, Ext) }

func readIntDir(dir string) (map[string]int, error) {
	names, err := readDirNames(dir)
	if err != nil {
		return nil, nil
	}
	out := map[string]int{}
	for _, n := range names {
		if !strings.HasSuffix(n, Ext) {
			continue
		}
		var v int
		if readLeaf(filepath.Join(dir, n), &v) {
			out[trimLeafExt(n)] = v
		}
	}
	return out, nil
}

func readStringDir(dir string) map[string]string {
	names, err := readDirNames(dir)
	if err != nil {
		return nil
	}
	out := map[string]string{}
	for _, n := range names {
		if !strings.HasSuffix(n, Ext) {
			continue
		}
		var v string
		if readLeaf(filepath.Join(dir, n), &v) {
			out[trimLeafExt(n)] = v
		}
	}
	return out
}

func readIntArrayDir(dir string) ([]int, error) {
	names, err := readDirNames(dir)
	if err != nil {
		return nil, nil
	}
	type item struct {
		idx int
		val int
	}
	items := make([]item, 0, len(names))
	for _, n := range names {
		if !strings.HasSuffix(n, Ext) {
			continue
		}
		idx, err := strconv.Atoi(trimLeafExt(n))
		if err != nil {
			return nil, fmt.Errorf("%s: element name %q is not an integer — an array directory is indexed by filename", dir, n)
		}
		var v int
		if readLeaf(filepath.Join(dir, n), &v) {
			items = append(items, item{idx: idx, val: v})
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].idx < items[j].idx })
	out := make([]int, 0, len(items))
	for _, it := range items {
		out = append(out, it.val)
	}
	return out, nil
}

func readNodeData(nodeDir string) (*NodeBuf.NodeData, error) {
	dataDir := filepath.Join(nodeDir, DirNodeData)
	if info, err := os.Stat(dataDir); err != nil || !info.IsDir() { // path-resolution-ok: the node's own data dir, not a scene path
		return nil, nil
	}

	var nd NodeBuf.NodeData
	readLeaf(filepath.Join(dataDir, FileLabel), &nd.Label)
	readLeaf(filepath.Join(dataDir, FileRepeat), &nd.Repeat)

	init, err := readIntArrayDir(filepath.Join(dataDir, DirInit))
	if err != nil {
		return nil, err
	}
	nd.Init = init

	state, err := readIntDir(filepath.Join(dataDir, DirState))
	if err != nil {
		return nil, err
	}
	if len(state) > 0 {
		nd.State = state
	}

	if rules := readStringDir(filepath.Join(dataDir, DirSendRule)); len(rules) > 0 {
		nd.SendRules = rules
	}

	return &nd, nil
}
