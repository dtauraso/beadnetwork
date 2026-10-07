package CardPanel

import "context"

type Inbox struct {
	Edits chan EditMsg
	Wake  chan struct{}
}

func NewInbox(depth int) Inbox {
	return Inbox{Edits: make(chan EditMsg, depth), Wake: make(chan struct{}, 1)}
}

type Inboxes struct {
	in map[string]Inbox
}

func (ib *Inboxes) Claim(id string, in Inbox) {
	if ib.in == nil {
		ib.in = map[string]Inbox{}
	}
	ib.in[id] = in
}

func (ib *Inboxes) Send(ctx context.Context, id string, msg EditMsg) bool {
	in, ok := ib.in[id]
	if !ok {
		return false
	}
	select {
	case in.Edits <- msg:
	case <-ctx.Done():
		return true
	}
	select {
	case in.Wake <- struct{}{}:
	default:
	}
	return true
}

func (ib *Inboxes) Broadcast(ctx context.Context, msg EditMsg) {
	for id := range ib.in {
		ib.Send(ctx, id, msg)
	}
}
