package Builder

import (
	"fmt"

	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
)

type NodeTiming struct {
	H int64
}

func (m *Model) TakeTiming(t Links.Timing) {
	if m.Timings == nil {
		m.Timings = map[string]NodeTiming{}
	}
	m.Timings[t.ID] = NodeTiming{H: t.H}
}

func (m Model) Behind() int {
	b := 0
	for _, id := range m.Nodes {
		if t, ok := m.Timings[id]; !ok || t.H < m.Now-1 {
			b++
		}
	}
	return b
}

func (m Model) TimeKey() string {
	return fmt.Sprintf("%d|%d", m.Now, m.Behind())
}
