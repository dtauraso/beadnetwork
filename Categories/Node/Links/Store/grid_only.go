package Store

import (
	"os"
	"path/filepath"

	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
)

var dragCopies = []string{"at-phi.bin", "at-theta.bin", "at-r.bin", "ref-phi.bin", "ref-theta.bin", "ref-s.bin", DirLinks}

func ShownPlace(nodeDir string) (Links.Place, Links.Ref, bool, error) {
	g, ok, err := ReadGrid(GridPath(nodeDir))
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
