package Builder

import (
	"path/filepath"

	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
)

const (
	DirCards   = "cards"
	DirMethods = "methods"
	FileItem   = "item.bin"
	FileCard   = "card.bin"
	MethodAtoB = "atob"
	MethodPair = "pair"
)

var MethodIDs = []string{MethodAtoB, MethodPair}

type Cards struct {
	Item    string
	Methods map[string]bool
	Card    int
}

func CardsDir(nodesRoot string) string { return filepath.Join(nodesRoot, DirCards) }

func ReadCards(dir string) Cards {
	c := Cards{Methods: map[string]bool{}, Card: 1}
	Links.ReadIfExists(filepath.Join(dir, FileItem), &c.Item)
	Links.ReadIfExists(filepath.Join(dir, FileCard), &c.Card)
	for _, m := range MethodIDs {
		var on bool
		Links.ReadIfExists(filepath.Join(dir, DirMethods, m+".bin"), &on)
		c.Methods[m] = on
	}
	return c
}

func WriteCards(dir string, c Cards) error {
	if err := Links.WriteAtomicIfChanged(filepath.Join(dir, FileItem), c.Item); err != nil {
		return err
	}
	if err := Links.WriteAtomicIfChanged(filepath.Join(dir, FileCard), c.Card); err != nil {
		return err
	}
	for _, m := range MethodIDs {
		if err := Links.WriteAtomicIfChanged(filepath.Join(dir, DirMethods, m+".bin"), c.Methods[m]); err != nil {
			return err
		}
	}
	return nil
}
