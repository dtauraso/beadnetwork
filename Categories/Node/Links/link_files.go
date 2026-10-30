package Links

import (
	"os"
	"path/filepath"
	"sort"
)

const (
	DirLinks = "links"
	DirDrag  = "drag"

	FileSource     = "source.bin"
	FileOrder      = "order.bin"
	FileStartPhi   = "start-phi.bin"
	FileStartTheta = "start-theta.bin"
	FileStartR     = "start-r.bin"
	FileDeltaPhi   = "delta-phi.bin"
	FileDeltaTheta = "delta-theta.bin"
	FileDeltaR     = "delta-r.bin"
	FileIncPhi     = "increment-phi.bin"
	FileIncTheta   = "increment-theta.bin"
	FileIncR       = "increment-r.bin"
	FileLockPhi    = "lock-phi.bin"
	FileLockTheta  = "lock-theta.bin"
	FileLocking    = "locking.bin"
	FilePicked     = "picked.bin"
)

func BaseLinksDir(nodeDir string) string { return filepath.Join(nodeDir, DirLinks) }

func DragLinksDir(nodeDir string) string { return filepath.Join(nodeDir, DirDrag, DirLinks) }

func linkName(from, to string) string { return from + "To" + to }

func ReadLinks(nodeDir, to string) ([]Link, error) {
	dir := DragLinksDir(nodeDir)
	if _, err := os.Stat(dir); err != nil { // path-resolution-ok: a node's edits replace its base links only once it has made one
		dir = BaseLinksDir(nodeDir)
	}
	return readLinksIn(dir, to)
}

func readLinksIn(linksDir, to string) ([]Link, error) {
	entries, err := os.ReadDir(linksDir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var all []Link
	for _, e := range entries {
		if !e.IsDir() { // path-resolution-ok: a link is a directory; skipping strays under links/
			continue
		}
		dir := filepath.Join(linksDir, e.Name())
		l := Link{To: to, Increment: Angles{Phi: 1, Theta: 1}, IncR: 1}
		if !ReadIfExists(filepath.Join(dir, FileSource), &l.From) {
			continue
		}
		ReadIfExists(filepath.Join(dir, FileOrder), &l.Order)
		ReadIfExists(filepath.Join(dir, FileStartPhi), &l.Start.Phi)
		ReadIfExists(filepath.Join(dir, FileStartTheta), &l.Start.Theta)
		ReadIfExists(filepath.Join(dir, FileStartR), &l.Start.R)
		ReadIfExists(filepath.Join(dir, FileDeltaPhi), &l.Delta.Phi)
		ReadIfExists(filepath.Join(dir, FileDeltaTheta), &l.Delta.Theta)
		ReadIfExists(filepath.Join(dir, FileDeltaR), &l.DeltaR)
		ReadIfExists(filepath.Join(dir, FileIncPhi), &l.Increment.Phi)
		ReadIfExists(filepath.Join(dir, FileIncTheta), &l.Increment.Theta)
		ReadIfExists(filepath.Join(dir, FileIncR), &l.IncR)
		ReadIfExists(filepath.Join(dir, FileLockPhi), &l.Lock.Phi)
		ReadIfExists(filepath.Join(dir, FileLockTheta), &l.Lock.Theta)
		ReadIfExists(filepath.Join(dir, FileLocking), &l.Locking)
		ReadIfExists(filepath.Join(dir, FilePicked), &l.Picked)
		all = append(all, l)
	}
	SortLinks(all)
	return all, nil
}

func SortLinks(ls []Link) {
	sort.SliceStable(ls, func(i, j int) bool {
		if ls[i].Order != ls[j].Order {
			return ls[i].Order < ls[j].Order
		}
		if ls[i].From != ls[j].From {
			return ls[i].From < ls[j].From
		}
		return ls[i].To < ls[j].To
	})
}

func WriteBaseLinks(nodeDir string, links []Link) error {
	return writeLinksIn(BaseLinksDir(nodeDir), links)
}

func WriteLinks(nodeDir string, links []Link) error {
	return writeLinksIn(DragLinksDir(nodeDir), links)
}

func RemoveLinksTo(nodeDir, from, to string) error {
	for _, dir := range []string{BaseLinksDir(nodeDir), DragLinksDir(nodeDir)} {
		if err := os.RemoveAll(filepath.Join(dir, linkName(from, to))); err != nil {
			return err
		}
	}
	return nil
}

func writeLinksIn(linksDir string, links []Link) error {
	keep := map[string]bool{}
	for _, l := range links {
		name := linkName(l.From, l.To)
		keep[name] = true
		dir := filepath.Join(linksDir, name)
		for file, v := range map[string]any{
			FileSource: l.From, FileOrder: l.Order,
			FileStartPhi: l.Start.Phi, FileStartTheta: l.Start.Theta, FileStartR: l.Start.R,
			FileDeltaPhi: l.Delta.Phi, FileDeltaTheta: l.Delta.Theta, FileDeltaR: l.DeltaR,
			FileIncPhi: l.Increment.Phi, FileIncTheta: l.Increment.Theta, FileIncR: l.IncR,
			FileLockPhi: l.Lock.Phi, FileLockTheta: l.Lock.Theta,
			FileLocking: l.Locking, FilePicked: l.Picked,
		} {
			if err := WriteAtomicIfChanged(filepath.Join(dir, file), v); err != nil {
				return err
			}
		}
	}
	if err := os.MkdirAll(linksDir, 0o755); err != nil {
		return err
	}
	entries, err := os.ReadDir(linksDir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if e.IsDir() && !keep[e.Name()] { // path-resolution-ok: removing a link the node no longer holds
			if err := os.RemoveAll(filepath.Join(linksDir, e.Name())); err != nil {
				return err
			}
		}
	}
	return nil
}
