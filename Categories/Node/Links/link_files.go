package Links

import (
	"os"
	"path/filepath"
	"sort"
)

const (
	DirLinks = "links"
	DirDrag  = "drag"

	FileTarget     = "target.bin"
	FileOrder      = "order.bin"
	FileStartPhi   = "start-phi.bin"
	FileStartTheta = "start-theta.bin"
	FileStartR     = "start-r.bin"
	FileDeltaPhi   = "delta-phi.bin"
	FileDeltaTheta = "delta-theta.bin"
	FileIncPhi     = "increment-phi.bin"
	FileIncTheta   = "increment-theta.bin"
	FileLockPhi    = "lock-phi.bin"
	FileLockTheta  = "lock-theta.bin"
	FileLocking    = "locking.bin"
	FilePicked     = "picked.bin"
)

func BaseLinksDir(nodeDir string) string { return filepath.Join(nodeDir, DirLinks) }

func DragLinksDir(nodeDir string) string { return filepath.Join(nodeDir, DirDrag, DirLinks) }

func linkName(from, to string) string { return from + "To" + to }

func ReadLinks(nodeDir, from string) ([]Link, error) {
	dir := DragLinksDir(nodeDir)
	if _, err := os.Stat(dir); err != nil { // path-resolution-ok: a node's edits replace its base links only once it has made one
		dir = BaseLinksDir(nodeDir)
	}
	return readLinksIn(dir, from)
}

func readLinksIn(linksDir, from string) ([]Link, error) {
	entries, err := os.ReadDir(linksDir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	type ordered struct {
		l     Link
		order int
	}
	var all []ordered
	for _, e := range entries {
		if !e.IsDir() { // path-resolution-ok: a link is a directory; skipping strays under links/
			continue
		}
		dir := filepath.Join(linksDir, e.Name())
		l := Link{From: from, Increment: Angles{Phi: 1, Theta: 1}}
		if !ReadIfExists(filepath.Join(dir, FileTarget), &l.To) {
			continue
		}
		var o int
		ReadIfExists(filepath.Join(dir, FileOrder), &o)
		ReadIfExists(filepath.Join(dir, FileStartPhi), &l.Start.Phi)
		ReadIfExists(filepath.Join(dir, FileStartTheta), &l.Start.Theta)
		ReadIfExists(filepath.Join(dir, FileStartR), &l.Start.R)
		ReadIfExists(filepath.Join(dir, FileDeltaPhi), &l.Delta.Phi)
		ReadIfExists(filepath.Join(dir, FileDeltaTheta), &l.Delta.Theta)
		ReadIfExists(filepath.Join(dir, FileIncPhi), &l.Increment.Phi)
		ReadIfExists(filepath.Join(dir, FileIncTheta), &l.Increment.Theta)
		ReadIfExists(filepath.Join(dir, FileLockPhi), &l.Lock.Phi)
		ReadIfExists(filepath.Join(dir, FileLockTheta), &l.Lock.Theta)
		ReadIfExists(filepath.Join(dir, FileLocking), &l.Locking)
		ReadIfExists(filepath.Join(dir, FilePicked), &l.Picked)
		all = append(all, ordered{l, o})
	}
	sort.SliceStable(all, func(i, j int) bool {
		if all[i].order != all[j].order {
			return all[i].order < all[j].order
		}
		return all[i].l.To < all[j].l.To
	})
	out := make([]Link, len(all))
	for i, o := range all {
		out[i] = o.l
	}
	return out, nil
}

func WriteBaseLinks(nodeDir string, links []Link) error {
	return writeLinksIn(BaseLinksDir(nodeDir), links)
}

func WriteLinks(nodeDir string, links []Link) error {
	return writeLinksIn(DragLinksDir(nodeDir), links)
}

func writeLinksIn(linksDir string, links []Link) error {
	keep := map[string]bool{}
	for i, l := range links {
		name := linkName(l.From, l.To)
		keep[name] = true
		dir := filepath.Join(linksDir, name)
		for file, v := range map[string]any{
			FileTarget: l.To, FileOrder: i,
			FileStartPhi: l.Start.Phi, FileStartTheta: l.Start.Theta, FileStartR: l.Start.R,
			FileDeltaPhi: l.Delta.Phi, FileDeltaTheta: l.Delta.Theta,
			FileIncPhi: l.Increment.Phi, FileIncTheta: l.Increment.Theta,
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
