package owners

import (
	interior "github.com/dtauraso/wirefold/Categories/Node/Interior"
	"github.com/dtauraso/wirefold/Categories/Node/nodegeom"
)

type Interior struct {
	stream  *interior.InteriorStream
	mailbox *interior.Mailbox

	lastCenter    Vec3
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
			o.stream.WriteEvents(snap.Events, interior.Vec3(center))
		} else {
			o.stream.WriteFull(snap.Present, snap.Value, snap.Ox, snap.Oy, snap.Oz, snap.Events, interior.Vec3(center))
		}
		wrote = true
	}

	if !wrote && o.hasLastCenter && nodegeom.Vec3(center) != nodegeom.Vec3(o.lastCenter) {
		o.stream.RewriteAtCenter(interior.Vec3(center))
	}
	o.lastCenter, o.hasLastCenter = Vec3(center), true
}
