package Startup

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Sheet/Scalar"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Pills/Builder"
	"github.com/dtauraso/beadnetwork/Categories/Node/Files"
	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
	"github.com/dtauraso/beadnetwork/Categories/Scene/Scenes"
)

type linkLayer struct {
	nodeDir string
	drag    bool
	links   []Links.Link
}

func FitTopStepsS(scenePath string) error {
	roots := []string{scenePath}
	for _, k := range Scenes.TopSteps(scenePath) {
		roots = append(roots, Scenes.TopStepDir(scenePath, k))
	}
	for _, root := range roots {
		if err := fitS(root); err != nil {
			return fmt.Errorf("Startup.FitTopStepsS: %s: %w", root, err)
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
	ids := subdirs(Files.NodesDir(root))
	for _, to := range ids {
		dirs = append(dirs, [2]string{Files.NodeDir(root, to), to})
	}
	steps := Builder.StepsDir(root)
	for _, step := range subdirs(steps) {
		nodes := filepath.Join(steps, step, Builder.DirNodes)
		for _, to := range subdirs(nodes) {
			dirs = append(dirs, [2]string{filepath.Join(nodes, to), to})
		}
	}
	var layers []*linkLayer
	for _, d := range dirs {
		ls, err := readLayers(d[0], d[1])
		if err != nil {
			return err
		}
		layers = append(layers, ls...)
	}
	sPath := Scenes.SFilePath(root)
	was := Scalar.Load(sPath, 1, 1)
	want := max(was, dropNodeS(root, ids))
	if want != was {
		if err := Scalar.Write(sPath, int32(want)); err != nil {
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
		drag, base := Files.PlaceDirs(Files.NodeDir(root, id))
		for _, path := range []string{filepath.Join(drag, Files.FileS), filepath.Join(base, Files.FileS)} {
			s := 1
			if Files.ReadIfExists(path, &s) {
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
