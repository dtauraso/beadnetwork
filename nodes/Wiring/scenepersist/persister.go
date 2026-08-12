package scenepersist

import "github.com/dtauraso/wirefold/nodes/Wiring/jsonpersist"

type Persister[T any] struct {
	Path  string
	Write func(path string, v T) error
	Tag   string
}

func (p *Persister[T]) Schedule(v T) {
	if p == nil || p.Path == "" {
		return
	}
	if err := p.Write(p.Path, v); err != nil {
		jsonpersist.LogPersistErr(p.Tag, p.Path, err)
	}
}
