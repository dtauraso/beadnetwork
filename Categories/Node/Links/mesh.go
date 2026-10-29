package Links

import (
	"context"
	"fmt"
)

const InboxDepth = 1024

type Report struct {
	ID      string
	S       int
	SNote   string
	At      Place
	Links   []Link
	Refused string
}

type Inbox struct {
	In   chan any
	Wake chan struct{}
}

type Mesh struct {
	in      map[string]Inbox
	order   []string
	reports map[string]chan Report
	timings map[string]chan Timing
}

func NewMesh() *Mesh {
	return &Mesh{in: map[string]Inbox{}, reports: map[string]chan Report{}, timings: map[string]chan Timing{}}
}

func (m *Mesh) Claim(id string) Inbox {
	if _, dup := m.in[id]; dup {
		panic(fmt.Sprintf("Links.Mesh.Claim: node %s claimed its links inbox twice — each node owns exactly one, made when the scene builds it", id))
	}
	in := Inbox{In: make(chan any, InboxDepth), Wake: make(chan struct{}, 1)}
	m.in[id] = in
	m.order = append(m.order, id)
	m.reports[id] = make(chan Report, 1)
	m.timings[id] = make(chan Timing, 1)
	return in
}

func (m *Mesh) IDs() []string { return m.order }

func (m *Mesh) Has(id string) bool {
	_, ok := m.in[id]
	return ok
}

func (m *Mesh) Send(ctx context.Context, from string, e Envelope) {
	if e.To != All {
		m.deliver(ctx, e.To, e.Msg)
		return
	}
	for _, id := range m.order {
		if id != from {
			m.deliver(ctx, id, e.Msg)
		}
	}
}

func (m *Mesh) deliver(ctx context.Context, to string, msg any) {
	in, ok := m.in[to]
	if !ok {
		return
	}
	select {
	case in.In <- msg:
	case <-ctx.Done():
		return
	}
	select {
	case in.Wake <- struct{}{}:
	default:
	}
}

func (m *Mesh) Report(r Report) {
	ch, ok := m.reports[r.ID]
	if !ok {
		return
	}
	select {
	case <-ch:
	default:
	}
	select {
	case ch <- r:
	default:
	}
}

func (m *Mesh) TakeReports() []Report {
	var out []Report
	for _, id := range m.order {
		select {
		case r := <-m.reports[id]:
			out = append(out, r)
		default:
		}
	}
	return out
}
