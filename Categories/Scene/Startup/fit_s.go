package Startup

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Sheet"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Pills/Builder"
	NodeBuf "github.com/dtauraso/beadnetwork/Categories/Node"
	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
	"github.com/dtauraso/beadnetwork/Categories/Scene/Scenes"
)

type linkLayer struct {
	nodeDir string
	drag    bool
	links   []Links.Link
}

func FitSequenceS(scenePath string) error {
	roots := []string{scenePath}
	for _, k := range Scenes.Sequences(scenePath) {
		roots = append(roots, Scenes.SequenceDir(scenePath, k))
	}
	for _, root := range roots {
		if err := fitS(root); err != nil {
			return fmt.Errorf("Startup.FitSequenceS: %s: %w", root, err)
		}
	}
	return nil
}

func subdirs(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var out []string
	for _, e := range entries {
		if e.IsDir() { // path-resolution-ok: nodes/ and the step clones hold one directory per node
			out = append(out, e.Name())
		}
	}
	return out
}

func readLayers(nodeDir, to string) ([]*linkLayer, error) {
	base, err := Links.ReadBaseLinks(nodeDir, to)
	if err != nil {
		return nil, err
	}
	drag, err := Links.ReadDragLinks(nodeDir, to)
	if err != nil {
		return nil, err
	}
	return []*linkLayer{{nodeDir, false, base}, {nodeDir, true, drag}}, nil
}

func fitS(root string) error {
	var dirs [][2]string
	ids := subdirs(NodeBuf.NodesDir(root))
	for _, to := range ids {
		dirs = append(dirs, [2]string{NodeBuf.NodeDir(root, to), to})
	}
	clones := filepath.Join(Builder.StepDir(root), Builder.DirStepClones)
	for _, step := range subdirs(clones) {
		for _, to := range subdirs(filepath.Join(clones, step)) {
			dirs = append(dirs, [2]string{filepath.Join(clones, step, to), to})
		}
	}
	var layers []*linkLayer
	need := 1
	for _, d := range dirs {
		ls, err := readLayers(d[0], d[1])
		if err != nil {
			return err
		}
		for _, ly := range ls {
			for _, l := range ly.links {
				need = Links.Lcm(need, l.Start.TicksNeeded())
			}
		}
		layers = append(layers, ls...)
	}
	sPath := Scenes.SFilePath(root)
	was := Sheet.LoadScalar(sPath, 1, 1)
	want := Links.Lcm(max(was, dropNodeS(root, ids)), need)
	if want != was {
		if err := Sheet.WriteScalar(sPath, int32(want)); err != nil {
			return err
		}
	}
	for _, ly := range layers {
		if err := ly.fit(want); err != nil {
			return err
		}
	}
	return nil
}

func dropNodeS(root string, ids []string) int {
	top := 1
	for _, id := range ids {
		drag, base := NodeBuf.PlaceDirs(NodeBuf.NodeDir(root, id))
		for _, path := range []string{filepath.Join(drag, NodeBuf.FileS), filepath.Join(base, NodeBuf.FileS)} {
			s := 1
			if NodeBuf.ReadIfExists(path, &s) {
				top = max(top, Links.ShownS(s))
				_ = os.Remove(path)
			}
		}
	}
	return top
}

func (ly *linkLayer) fit(want int) error {
	changed := false
	for i := range ly.links {
		l := &ly.links[i]
		if l.Start.S != want {
			l.SetS(want)
			changed = true
		}
	}
	if !changed {
		return nil
	}
	if ly.drag {
		return Links.WriteLinks(ly.nodeDir, ly.links)
	}
	return Links.WriteBaseLinks(ly.nodeDir, ly.links)
}
