package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Sheet"
	NodeBuf "github.com/dtauraso/beadnetwork/Categories/Node"
	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
	SceneBuf "github.com/dtauraso/beadnetwork/Categories/Scene"
	"github.com/dtauraso/beadnetwork/Categories/Scene/Scenes"
	"github.com/dtauraso/beadnetwork/Categories/Scene/Topology"
	"github.com/dtauraso/beadnetwork/Categories/Vectors/polarindex"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: readscene <scene dir> [sequence ...]")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Prints the values a scene's sequences are saved as — the frame,")
		fmt.Fprintln(os.Stderr, "each node's place and s, each link's start, Δ, increment and lock —")
		fmt.Fprintln(os.Stderr, "as text, with the x, y, z the simulation makes from them. Run it")
		fmt.Fprintln(os.Stderr, "before and after pressing a button and diff the two.")
		os.Exit(2)
	}
	root := os.Args[1]
	seqs := Scenes.Sequences(root)
	if len(os.Args) > 2 {
		seqs = nil
		for _, a := range os.Args[2:] {
			k, err := strconv.Atoi(a)
			if err != nil {
				fmt.Fprintf(os.Stderr, "readscene: sequence %q is not a number\n", a)
				os.Exit(2)
			}
			seqs = append(seqs, k)
		}
	}

	constants, err := SceneBuf.LoadSceneFrame(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "readscene: %v\n", err)
		os.Exit(1)
	}
	sceneS := Sheet.LoadScalar(Scenes.SFilePath(root), 1, 1)

	out := bufio.NewWriter(os.Stdout)
	status := 0
	for _, k := range seqs {
		if err := dumpSequence(out, root, k, constants, sceneS); err != nil {
			fmt.Fprintf(os.Stderr, "readscene: sequence %d: %v\n", k, err)
			status = 1
		}
	}
	if err := out.Flush(); err != nil {
		fmt.Fprintf(os.Stderr, "readscene: flush: %v\n", err)
		status = 1
	}
	os.Exit(status)
}

func readPlace(dir string) *polarindex.Place {
	var p polarindex.Place
	if !Links.ReadIfExists(filepath.Join(dir, NodeBuf.FileAtPhi), &p.Phi) ||
		!Links.ReadIfExists(filepath.Join(dir, NodeBuf.FileAtTheta), &p.Theta) ||
		!Links.ReadIfExists(filepath.Join(dir, NodeBuf.FileAtR), &p.R) {
		return nil
	}
	return &p
}

func readInt(path string) *int {
	var v int
	if !Links.ReadIfExists(path, &v) {
		return nil
	}
	return &v
}

func loadSequence(root string, k int, constants polarindex.SceneConstants) (Topology.TopoSpec, error) {
	spec := Topology.TopoSpec{Constants: constants, Rho: SceneBuf.LoadRho(root)}
	seqDir := Scenes.SequenceDir(root, k)
	nodesDir := NodeBuf.NodeDir(seqDir, "")
	entries, err := os.ReadDir(nodesDir)
	if err != nil {
		return spec, err
	}
	var ids []string
	for _, e := range entries {
		if e.IsDir() { // path-resolution-ok: a node is a numbered directory under nodes/
			ids = append(ids, e.Name())
		}
	}
	sort.Slice(ids, func(i, j int) bool {
		a, _ := strconv.Atoi(ids[i])
		b, _ := strconv.Atoi(ids[j])
		return a < b
	})
	for _, id := range ids {
		nodeDir := NodeBuf.NodeDir(seqDir, id)
		drag, base := NodeBuf.PlaceDirs(nodeDir)
		n := Topology.Node{ID: id}
		Links.ReadIfExists(filepath.Join(base, NodeBuf.FileType), &n.Type)
		n.At = readPlace(base)
		if at := readPlace(drag); at != nil {
			n.At = at
		}
		n.S = readInt(filepath.Join(base, NodeBuf.FileS))
		if s := readInt(filepath.Join(drag, NodeBuf.FileS)); s != nil {
			n.S = s
		}
		spec.Nodes = append(spec.Nodes, n)
		links, err := Links.ReadLinks(nodeDir, id)
		if err != nil {
			return spec, fmt.Errorf("node %s links: %w", id, err)
		}
		spec.Links = append(spec.Links, links...)
	}
	return spec, nil
}

func dumpSequence(out *bufio.Writer, root string, k int, constants polarindex.SceneConstants, sceneS int) error {
	spec, err := loadSequence(root, k, constants)
	if err != nil {
		return err
	}
	seed := Topology.LinkSeed(spec, sceneS, 0)
	f := seed.Frame

	fmt.Fprintf(out, "sequence %d  %q\n", k, Scenes.SequenceName(root, k))
	fmt.Fprintf(out, "  frame  rho=%s  constant-r=%s  max-index-phi=%d  max-index-theta=%d  turn=12·s ticks\n",
		num(f.Rho), num(f.Cr), f.Tphi, f.Ttheta)

	for _, n := range spec.Nodes {
		p, placed := seed.Places[n.ID]
		if !placed {
			fmt.Fprintf(out, "  node %-3s s=%d  (no place)\n", n.ID, seed.Scales[n.ID])
			continue
		}
		v := f.At(p)
		fmt.Fprintf(out, "  node %-3s s=%d  place φ=%s θ=%s r=%s   x=%s y=%s z=%s\n",
			n.ID, seed.Scales[n.ID], num(p.Phi), num(p.Theta), num(p.R), num(v.X), num(v.Y), num(v.Z))
	}

	for _, l := range spec.Links {
		s := seed.Scales[l.To]
		d := l.Dir(s)
		v := f.Vec(s, d)
		fmt.Fprintf(out, "  link %s→%s  start φ=%s θ=%s r=%s   Δ φ=%d θ=%d r=%d   start+Δ φ=%s θ=%s r=%s   vector x=%s y=%s z=%s\n",
			l.From, l.To, num(l.Start.Phi), num(l.Start.Theta), num(l.Start.R),
			l.Delta.Phi, l.Delta.Theta, l.Delta.R, num(d.Phi), num(d.Theta), num(d.R), num(v.X), num(v.Y), num(v.Z))
		fmt.Fprintf(out, "           increment φ=%d θ=%d r=%d   lock φ=%d θ=%d   locking=%v  picked=%v  order=%d\n",
			l.Increment.Phi, l.Increment.Theta, l.Increment.R, l.Lock.Phi, l.Lock.Theta, l.Locking, l.Picked, l.Order)
		if from, ok := seed.Places[l.From]; ok {
			if to, ok := seed.Places[l.To]; ok {
				end := f.At(from).Add(v)
				got := f.At(to)
				fmt.Fprintf(out, "           %s + vector = x=%s y=%s z=%s   node %s is at x=%s y=%s z=%s\n",
					l.From, num(end.X), num(end.Y), num(end.Z), l.To, num(got.X), num(got.Y), num(got.Z))
			}
		}
	}
	fmt.Fprintln(out)
	return nil
}

func num(x float64) string {
	if x > -5e-7 && x < 5e-7 {
		x = 0
	}
	return strconv.FormatFloat(x, 'f', 6, 64)
}
