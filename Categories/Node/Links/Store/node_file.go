package Store

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
)

const (
	FileNode    = "node.bin"
	nodeMagic   = "NODE"
	nodeVersion = 1
)

func NodePath(nodeDir string) string { return filepath.Join(nodeDir, FileNode) }

func WriteNode(path string, n Links.Node) error {
	var b bytes.Buffer
	b.WriteString(nodeMagic)
	le := binary.LittleEndian
	_ = binary.Write(&b, le, uint32(nodeVersion))
	putStr(&b, n.ID)
	rows := colRows([]Col{colOf(n)})
	_ = binary.Write(&b, le, uint32(len(rows)))
	for _, r := range rows {
		putRow(&b, r)
	}
	return writeIfChanged(path, b.Bytes())
}

func ReadNodeRows(path string) (string, []GridRow, error) {
	return readOneCol(path, nodeMagic, nodeVersion)
}

func readOneCol(path, magic string, want uint32) (string, []GridRow, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", nil, err
	}
	if !strings.HasPrefix(string(data), magic) {
		return "", nil, fmt.Errorf("Store.ReadNodeRows: %s does not start with %q, so it is not a node file", path, magic)
	}
	r, le := bytes.NewReader(data[len(magic):]), binary.LittleEndian
	var ver, nrows uint32
	_ = binary.Read(r, le, &ver)
	if ver != want {
		return "", nil, fmt.Errorf("Store.ReadNodeRows: %s is %s version %d, and this build reads version %d", path, magic, ver, want)
	}
	id, _ := getStr(r)
	if err := binary.Read(r, le, &nrows); err != nil {
		return id, nil, fmt.Errorf("Store.ReadNodeRows: %s ends inside its header: %w", path, err)
	}
	rows := make([]GridRow, nrows)
	for i := range rows {
		if rows[i], err = getRow(r, typedVersion, 1); err != nil {
			return id, nil, fmt.Errorf("Store.ReadNodeRows: %s ends inside row %d (%s %s): %w", path, i, rows[i].Band, rows[i].Name, err)
		}
	}
	return id, rows, nil
}

func ReadNode(path string) (Links.Node, bool, error) {
	return readNode(path, nodeMagic, nodeVersion)
}

func ReadOneColGrid(path string) (Links.Node, bool, error) {
	return readNode(path, gridMagic, oneColGridVersion)
}

func readNode(path, magic string, ver uint32) (Links.Node, bool, error) {
	id, rows, err := readOneCol(path, magic, ver)
	if errors.Is(err, fs.ErrNotExist) {
		return Links.Node{}, false, nil
	}
	if err != nil {
		return Links.Node{}, false, err
	}
	cols := make([]Col, 1)
	if err := colsOf(id, rows, cols); err != nil {
		return Links.Node{}, false, fmt.Errorf("Store.ReadNode: %s %w", path, err)
	}
	return cols[0].node(id), true, nil
}

var dragCopies = []string{"at-phi.bin", "at-theta.bin", "at-r.bin", "ref-phi.bin", "ref-theta.bin", "ref-s.bin", DirLinks}

func ShownPlace(nodeDir string) (Links.Place, Links.Ref, bool, error) {
	if n, ok, err := ReadNode(NodePath(nodeDir)); err != nil || ok {
		return n.At, n.Ref, ok, err
	}
	g, ok, err := ReadOldGrid(GridPath(nodeDir))
	if err != nil || !ok {
		return Links.Place{}, Links.Ref{}, false, err
	}
	c := g.ShownCol()
	return c.At, c.Ref, true, nil
}

func RemoveDragCopies(nodeDir string) error {
	for _, name := range dragCopies {
		if err := os.RemoveAll(filepath.Join(nodeDir, DirDrag, name)); err != nil {
			return err
		}
	}
	return nil
}

func removeFromNode(nodeDir, from string) error {
	path := NodePath(nodeDir)
	n, ok, err := ReadNode(path)
	if err != nil || !ok {
		return err
	}
	n.In = slices.DeleteFunc(n.In, func(in Links.InCell) bool { return in.Link.From == from })
	return WriteNode(path, n)
}
