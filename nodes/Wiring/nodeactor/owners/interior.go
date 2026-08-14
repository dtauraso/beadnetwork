package owners

import "github.com/dtauraso/wirefold/nodes/Wiring/interior"

type Interior struct {
	stream  *interior.InteriorStream
	mailbox *interior.Mailbox
}

func (o *Interior) SetInteriorStream(stream *interior.InteriorStream, mailbox *interior.Mailbox) {
	o.stream = stream
	o.mailbox = mailbox
}

func (o *Interior) WriteFrames() {
	if o.stream == nil || o.mailbox == nil {
		return
	}
	for {
		snap, ok := o.mailbox.TryRecv()
		if !ok {
			return
		}
		if snap.EventsOnly {
			o.stream.WriteEvents(snap.Events)
		} else {
			o.stream.WriteFull(snap.Present, snap.Value, snap.Ox, snap.Oy, snap.Oz, snap.Events)
		}
	}
}
