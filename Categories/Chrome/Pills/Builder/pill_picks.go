package Builder

import "github.com/dtauraso/beadnetwork/Categories/Node/Links"

const OpIncrement = "increment"

func (m Model) Picked() []Links.Link {
	var out []Links.Link
	for _, l := range m.allLinks() {
		if l.Picked {
			out = append(out, l)
		}
	}
	return out
}

const (
	OpAllBoth  = "all-both"
	OpAllLock  = "all-lock"
	OpAllLocks = "all-locks"
	OpAllInc   = "all-increment"
	OpAllOne   = "all-one"
)
