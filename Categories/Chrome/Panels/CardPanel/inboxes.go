package CardPanel

import "context"

type Inboxes struct {
	in map[string]chan EditMsg
}

func (ib *Inboxes) Claim(id string, ch chan EditMsg) {
	if ib.in == nil {
		ib.in = map[string]chan EditMsg{}
	}
	ib.in[id] = ch
}

func (ib *Inboxes) Send(ctx context.Context, id string, msg EditMsg) bool {
	ch, ok := ib.in[id]
	if !ok {
		return false
	}
	select {
	case ch <- msg:
	case <-ctx.Done():
	}
	return true
}

func (ib *Inboxes) Broadcast(ctx context.Context, msg EditMsg) {
	for id := range ib.in {
		ib.Send(ctx, id, msg)
	}
}
