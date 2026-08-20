package owners

import (
	"github.com/dtauraso/wirefold/src/Node/Interior"
	"github.com/dtauraso/wirefold/src/Node/Wiring/nodegeom"
	"github.com/dtauraso/wirefold/src/spatial"
)

type Interior struct {
	stream  *interior.InteriorStream
	mailbox *interior.Mailbox

	lastCenter    spatial.Vec3
	hasLastCenter bool
}

func (o *Interior) SetInteriorStream(stream *interior.InteriorStream, mailbox *interior.Mailbox) {
	o.stream = stream
	o.mailbox = mailbox
}

func (o *Interior) WriteFrames(self nodegeom.NodeGeom) {
	if o.stream == nil || o.mailbox == nil {
		return
	}
	center := nodegeom.NodeWorldPos(self)
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
