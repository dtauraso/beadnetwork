package Store

import (
	"os"
	"path/filepath"
	"strconv"

	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
)

func StepDir(levelDir string, step int) string {
	return filepath.Join(levelDir, "steps", strconv.Itoa(step))
}

func StepNumbers(levelDir string) []int {
	var out []int
	entries, _ := os.ReadDir(filepath.Join(levelDir, "steps"))
	for _, e := range entries {
		if n, err := strconv.Atoi(e.Name()); e.IsDir() && err == nil && n > 0 { // path-resolution-ok: steps/ holds one numbered directory per step; skipping strays
			out = append(out, n)
		}
	}
	if len(out) == 0 {
		out = []int{1}
	}
	return out
}

func GridPath(nodeDir string) string { return filepath.Join(nodeDir, FileGrid) }

func ConvertGrid(level, id string, stepNodeDirs map[int]string, live []Links.Link, at Links.Place, ref Links.Ref, s int) (Links.Grid, error) {
	g := Links.Grid{ID: id, Level: level}
	for n := range stepNodeDirs {
		g.Cols = append(g.Cols, Links.Col{Step: n})
	}
	sortCols(g.Cols)
	for i := range g.Cols {
		n := g.Cols[i].Step
		c := Links.FreshCol(n, live, at, ref)
		old, err := ReadBaseLinks(stepNodeDirs[n], id)
		if err != nil {
			return g, err
		}
		for _, o := range old {
			for j := range c.In {
				if c.In[j].Link.From == o.From {
					c.In[j].Link = Links.Normalized(o, s)
				}
			}
		}
		g.Cols[i] = c
	}
	g.Shown = g.Cols[0].Step
	return g, nil
}

func sortCols(cols []Links.Col) {
	for i := 1; i < len(cols); i++ {
		for j := i; j > 0 && cols[j].Step < cols[j-1].Step; j-- {
			cols[j], cols[j-1] = cols[j-1], cols[j]
		}
	}
}

func RemoveOldLinks(nodeDirs ...string) error {
	for _, d := range nodeDirs {
		if err := os.RemoveAll(BaseLinksDir(d)); err != nil {
			return err
		}
	}
	return nil
}
