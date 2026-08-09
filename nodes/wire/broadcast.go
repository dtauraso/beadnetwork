// broadcast.go — ONE JOB: emitting the SAME value onto a SET of Outs in one call.
// Broadcast is the set; PlaceDrivenAllAt is the single emission across it, and the
// reason it exists as its own operation rather than a caller-side loop is the ONE
// shared tick reading it passes to every Out. Each Out's own placement is
// drive_item.go / out_port.go.

package wire

// Broadcast is a broadcast port: a slice of Outs the node emits the same
// value onto, each its own independent 1:1 wire.
type Broadcast []*Out

// PlaceDrivenAllAt places value v (no walker) on EVERY Out in the set, emitting
// the SendWire trace for each and appending a DriveItem per Out to dst. tick is
// the CALLER's own clock reading, read ONCE and passed to every Out in the
// set — that single shared reading is what guarantees every bead of this one
// broadcast emission shares the same placementTick (the bug this replaced: an
// earlier version let each wire's own drain pass stamp its own tick, so a
// broadcast could straddle a tick boundary between two of its beads). Once
// placed, each wire's own driver (its source node's mover) advances and
// delivers its bead independently, so the traversal still animates
// concurrently. Chan-mode Outs send immediately and contribute inert items.
func (outs Broadcast) PlaceDrivenAllAt(v int, dst []DriveItem, tick int64) []DriveItem {
	for _, o := range outs {
		if o == nil {
			continue
		}
		dst = append(dst, o.PlaceDrivenAt(v, tick))
	}
	return dst
}
