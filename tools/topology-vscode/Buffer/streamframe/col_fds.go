package streamframe

import (
	"fmt"
	"os"

	B "github.com/dtauraso/wirefold/tools/topology-vscode/Buffer"
	"github.com/dtauraso/wirefold/tools/topology-vscode/Buffer/colstream"
)

const StreamKindCol = "col"

type ColumnStreams struct {
	base  int
	nodes int
	edges int
	on    bool
}

const streamKindColCount = "colcount"

func NewColumnStreams(fds StreamFDs, nodes, edges int) ColumnStreams {
	base, ok := fds[StreamKindCol]
	if !ok {
		return ColumnStreams{}
	}

	want := B.ColumnStreamCount(nodes, edges)
	if got, ok := fds[streamKindColCount]; ok && got != want {
		panic(fmt.Sprintf(
			"column-stream layout skew: the extension host allocated %d column pipes for %d nodes / "+
				"%d edges, this binary's schema wants %d. The two were built from different "+
				"layouts, so every column would land on the wrong fd. Run \"Developer: Reload "+
				"Window\" — reopening the file reloads only the webview, not the extension host "+
				"that allocates these pipes.",
			got, nodes, edges, want))
	}
	return ColumnStreams{base: base, nodes: nodes, edges: edges, on: true}
}

func (c ColumnStreams) On() bool { return c.on }

func (c ColumnStreams) singletonBase() int {
	return c.base
}

func (c ColumnStreams) nodeBase(row int) int {
	return c.base + B.ColumnsInSingletonStreams + row*B.ColumnsPerNodeStream
}

func (c ColumnStreams) edgeBase(row int) int {
	return c.base + B.ColumnsInSingletonStreams + c.nodes*B.ColumnsPerNodeStream + row*B.ColumnsPerEdgeStream
}

func (c ColumnStreams) open(base, col int, what string) *os.File {
	fd := base + col
	return os.NewFile(uintptr(fd), fmt.Sprintf("%s-col%d-fd%d", what, col, fd))
}

func (c ColumnStreams) NodeColumns(row int) *colstream.ColumnSet {
	if !c.on || row < 0 || row >= c.nodes {
		return nil
	}
	set := colstream.NewColumnSet(B.ColumnsPerNodeStream)
	base := c.nodeBase(row)
	for i := 0; i < B.ColumnsPerNodeStream; i++ {
		set.Attach(i, c.open(base, i, fmt.Sprintf("node%d", row)))
	}
	return set
}

func (c ColumnStreams) EdgeColumns(row int) *colstream.ColumnSet {
	if !c.on || row < 0 || row >= c.edges {
		return nil
	}
	set := colstream.NewColumnSet(B.ColumnsPerEdgeStream)
	base := c.edgeBase(row)
	for i := 0; i < B.ColumnsPerEdgeStream; i++ {
		set.Attach(i, c.open(base, i, fmt.Sprintf("edge%d", row)))
	}
	return set
}

func (c ColumnStreams) SingletonColumns() *colstream.ColumnSet {
	if !c.on {
		return nil
	}
	set := colstream.NewColumnSet(B.ColumnsInSingletonStreams)
	base := c.singletonBase()
	for i := 0; i < B.ColumnsInSingletonStreams; i++ {
		set.Attach(i, c.open(base, i, "view"))
	}
	return set
}
