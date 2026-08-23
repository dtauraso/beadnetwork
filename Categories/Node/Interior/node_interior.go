package interior

import ()

type Interior struct {
	stream  *InteriorStream
	mailbox *Mailbox

	lastCenter    Vec3
	hasLastCenter bool
}

func (o *Interior) SetInteriorStream(stream *InteriorStream, mailbox *Mailbox) {
	o.stream = stream
	o.mailbox = mailbox
}

func (o *Interior) WriteFrames(center Vec3) {
	if o.stream == nil || o.mailbox == nil {
		return
	}
	wrote := false
	for {
		snap, ok := o.mailbox.TryRecv()
		if !ok {
			break
		}
		if snap.EventsOnly {
			o.stream.WriteEvents(snap.Events, center)
		} else {
			o.stream.WriteFull(snap.Present, snap.Value, snap.Ox, snap.Oy, snap.Oz, snap.Events, center)
		}
		wrote = true
	}

	if !wrote && o.hasLastCenter && center != o.lastCenter {
		o.stream.RewriteAtCenter(center)
	}
	o.lastCenter, o.hasLastCenter = center, true
}
