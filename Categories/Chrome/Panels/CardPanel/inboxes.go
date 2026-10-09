package CardPanel

import "context"

type Inbox struct {
	Edits   chan EditMsg
	Steps   chan struct{}
	Resets  chan struct{}
	Starts  chan struct{}
	Centres chan string
	Wake    chan struct{}
}

func NewInbox(depth int) Inbox {
	return Inbox{Edits: make(chan EditMsg, depth), Steps: make(chan struct{}, depth), Resets: make(chan struct{}, 1),
		Starts: make(chan struct{}, 1), Centres: make(chan string, 1), Wake: make(chan struct{}, 1)}
}

func (ib *Inboxes) SetCentre(id string) {
	for _, in := range ib.in {
		select {
		case <-in.Centres:
		default:
		}
		select {
		case in.Centres <- id:
		default:
		}
		wake(in)
	}
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
	wake(in)
	return true
}

func (ib *Inboxes) Broadcast(ctx context.Context, msg EditMsg) {
	for id := range ib.in {
		ib.Send(ctx, id, msg)
	}
}

func (ib *Inboxes) Step() {
	for _, in := range ib.in {
		select {
		case in.Steps <- struct{}{}:
		default:
		}
		wake(in)
	}
}

func (ib *Inboxes) Reset() {
	for _, in := range ib.in {
		select {
		case in.Resets <- struct{}{}:
		default:
		}
		wake(in)
	}
}

func (ib *Inboxes) Start() {
	for _, in := range ib.in {
		select {
		case in.Starts <- struct{}{}:
		default:
		}
		wake(in)
	}
}

func wake(in Inbox) {
	select {
	case in.Wake <- struct{}{}:
	default:
	}
}
