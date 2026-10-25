package Builder

import (
	"fmt"
	"math"
	"strconv"

	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
)

type NodeTiming struct {
	H, K int64
	N    float64
}

func (m *Model) TakeTiming(t Links.Timing) {
	if m.Timings == nil {
		m.Timings = map[string]NodeTiming{}
	}
	prev, seen := m.Timings[t.ID]
	nt := NodeTiming{H: t.H, K: t.K, N: prev.N}
	if seen && t.H > prev.H && t.K >= prev.K {
		n := float64(t.K-prev.K) / float64(t.H-prev.H)
		if prev.N == 0 {
			nt.N = n
		} else {
			nt.N = 0.8*prev.N + 0.2*n
		}
	}
	m.Timings[t.ID] = nt
}

func (m Model) NBar() float64 {
	sum, c := 0.0, 0
	for _, id := range m.Nodes {
		if t, ok := m.Timings[id]; ok && t.N > 0 {
			sum += t.N
			c++
		}
	}
	if c == 0 {
		return 0
	}
	return sum / float64(c)
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

func (m Model) NOf(id string) string {
	t, ok := m.Timings[id]
	if !ok || t.N == 0 {
		return "—"
	}
	return Sig2(t.N)
}

func (m Model) MachinePerSec() string {
	if m.TickMs < 1 {
		return "—"
	}
	n := m.NBar()
	if n == 0 {
		return "—"
	}
	return Sig2(n*1000/float64(m.TickMs)/1e6) + "M"
}

func (m Model) TimeKey() string {
	return fmt.Sprintf("%d|%s|%s|%d|%v", m.Now, Sig2(m.NBar()), m.MachinePerSec(), m.Behind(), m.nKeys())
}

func (m Model) nKeys() []string {
	out := make([]string, len(m.Nodes))
	for i, id := range m.Nodes {
		out[i] = m.NOf(id)
	}
	return out
}

func Sig2(x float64) string {
	if x == 0 || math.IsNaN(x) || math.IsInf(x, 0) {
		return "0"
	}
	return strconv.FormatFloat(x, 'g', 2, 64)
}
