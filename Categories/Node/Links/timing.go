package Links

type Timing struct {
	ID string
	H  int64
	K  int64
}

func (m *Mesh) PostTiming(t Timing) {
	ch, ok := m.timings[t.ID]
	if !ok {
		return
	}
	select {
	case <-ch:
	default:
	}
	select {
	case ch <- t:
	default:
	}
}

func (m *Mesh) TakeTimings() []Timing {
	var out []Timing
	for _, id := range m.order {
		select {
		case t := <-m.timings[id]:
			out = append(out, t)
		default:
		}
	}
	return out
}
