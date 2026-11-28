package main

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Pills/Builder/StepStore"
	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
	"github.com/dtauraso/beadnetwork/Categories/Node/Links/Store"
	"github.com/dtauraso/beadnetwork/Categories/Scene/Scenes"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: go run ./scripts/migrate-grids <scene dir>...")
		os.Exit(2)
	}
	for _, scene := range os.Args[1:] {
		for _, path := range Store.OldGrids(Scenes.StepRoots(scene)) {
			if err := migrate(path); err != nil {
				fmt.Fprintf(os.Stderr, "migrate-grids: %s: %v\n", path, err)
				os.Exit(1)
			}
		}
	}
}

func migrate(path string) error {
	nodeDir := filepath.Dir(path)
	root := filepath.Dir(filepath.Dir(nodeDir))
	if Store.IsOneColGrid(path) {
		return oneCol(path, nodeDir)
	}
	g, ok, err := Store.ReadOldGrid(path)
	if err != nil || !ok {
		return err
	}
	leaves := StepStore.Leaves(root, nil)
	isLeaf := func(p []int) bool {
		return slices.ContainsFunc(leaves, func(l []int) bool { return slices.Equal(l, p) })
	}
	level := StepStore.LevelPath(g.Level)
	at := func(step int) []int { return append(slices.Clone(level), step) }
	shown := g.ShownCol()
	sums := map[string]Links.Ints{}
	for _, c := range g.Cols {
		p := at(c.Step)
		if !isLeaf(p) {
			for _, in := range c.In {
				fmt.Printf("migrate-grids: %s: step %s holds inner steps, so %s→%s's amount %+v there is dropped\n", path, StepStore.PathKey(p), in.Link.From, g.ID, in.Link.Delta)
			}
			continue
		}
		for _, in := range c.In {
			l := in.Link
			a := Links.Ints{Phi: l.Delta.Phi, Theta: l.Delta.Theta, R: l.Delta.R}
			if err := StepStore.WriteAmount(StepStore.PathDir(root, p), StepStore.Edge{From: l.From, To: g.ID}, StepStore.Amount{D: a, Inc: l.Increment, Picked: l.Picked}); err != nil {
				return err
			}
			if l.Picked && slices.Compare(p, at(shown.Step)) <= 0 {
				sums[l.From] = sums[l.From].Add(a)
			}
		}
	}
	n := Links.Node{ID: g.ID, Ref: shown.Ref, At: shown.At, Pos: shown.Pos, Label: shown.Label, Out: shown.Out}
	for _, in := range shown.In {
		l, b := in.Link, g.Base[in.Link.From]
		d0 := b.Total.Add(l.Delta)
		sum := sums[l.From]
		l.Start = b.Origin
		l.Start.Less(sum.Sub(d0))
		l.Delta = Links.Ints{Phi: sum.Phi, Theta: sum.Theta, R: sum.R, S: l.Delta.S}
		l.Picked = false
		in.Link = l
		n.In = append(n.In, in)
	}
	if err := Store.WriteNode(Store.NodePath(nodeDir), n); err != nil {
		return err
	}
	return done(path, nodeDir, n)
}

func oneCol(path, nodeDir string) error {
	n, ok, err := Store.ReadOneColGrid(path)
	if err != nil || !ok {
		return err
	}
	for i := range n.In {
		l := &n.In[i].Link
		if !l.Delta.Zero() {
			fmt.Printf("migrate-grids: %s: %s→%s's Δ %+v has no step to hold it, so it moves into START and the Δ sum starts at 0\n", path, l.From, n.ID, l.Delta)
		}
		l.Start.Less(Links.Ints{}.Sub(l.Delta))
		l.Delta = Links.Ints{S: l.Delta.S}
		l.Picked = false
	}
	return done(path, nodeDir, n)
}

func done(path, nodeDir string, n Links.Node) error {
	if err := Store.WriteNode(Store.NodePath(nodeDir), n); err != nil {
		return err
	}
	fmt.Printf("migrate-grids: %s: node %s keeps %d links in %s; the grid is kept as %s.old\n", path, n.ID, len(n.In), Store.FileNode, Store.FileGrid)
	return os.Rename(path, path+".old")
}
