// breadcrumb_labels.go — the DEBUG BREADCRUMB sub-vocabulary (Kind==KindBreadcrumb rows
// only): the BreadcrumbLabel* enum (Buffer/layout.go's bufLayoutEvent.Label column), its
// string-name table (mirrored into TS for the .probe decode/log by gen-node-defs), and the
// name→id resolver every Breadcrumb() call site's structured buffer emit goes through.
// Split out of Trace.go for the same one-concern-per-file reason as kind_events.go.
package Trace

// BreadcrumbLabel* enumerate the breadcrumb call sites (Buffer/layout.go's
// bufLayoutEvent.Label column, Kind==KindBreadcrumb rows only). Order is the wire id —
// append only; do not reorder or delete a label without a migration. BreadcrumbLabels
// is the string lookup gen-node-defs mirrors into TS for the .probe decode/log.
const (
	BreadcrumbTopologyLoaded uint8 = iota
	BreadcrumbRowSeedCountMismatch
	BreadcrumbPoleToggleGo
	BreadcrumbWindowClear
	BreadcrumbWindowOpen
	BreadcrumbDwellStart
	BreadcrumbAbcDrag
	BreadcrumbWireSendBufferFull
	// BreadcrumbDragCommit reports a node's own drag commit (owner-goroutine handle's
	// moveMsgKindDrag case) — the new position this node's own goroutine just committed.
	BreadcrumbDragCommit
	// BreadcrumbWireBreadcrumbsDropped reports how many KindBreadcrumb rows
	// PacedWire.Send's non-blocking breadcrumbCh send silently dropped since
	// the last report (breadcrumbCh's own doc comment, paced_wire.go) — Value
	// carries the dropped count. Emitted once room reappears on breadcrumbCh,
	// so the diagnostic channel's own lossiness is never itself silent.
	BreadcrumbWireBreadcrumbsDropped
	// BreadcrumbChainAim: diagnostic-only (task/log-node4-chain-aim), one per outgoing
	// target per chainBeads() call — see chain_beads.go's tr.Breadcrumb("chain-aim", ...)
	// call site for the exact fields packed into Text.
	BreadcrumbChainAim
	// BreadcrumbNeighborCenterRecv: diagnostic-only (task/log-node4-chain-aim), fired in
	// nodeMover.handle's moveMsgKindNeighborCenter case — records that a neighbor-center
	// push arrived (sender id + pushed center).
	BreadcrumbNeighborCenterRecv
	// BreadcrumbNeighborSetCRecv: diagnostic-only (task/log-node4-chain-aim), fired in
	// nodeMover.handle's moveMsgKindNeighborSetC case — records that a neighbor-setC
	// (edge re-quantize) message arrived (sender id).
	BreadcrumbNeighborSetCRecv
	// BreadcrumbBeadCrud: diagnostic-only (task/log-node2-bead-crud), one per commitNodeMoveLocal
	// call — the dragged node's own event plus every touching bead's full CRUD arithmetic
	// (why each returned none/add/remove), packed into Text by quantized_move.go.
	BreadcrumbBeadCrud
	// BreadcrumbPairSeedUnknown: one at BUILD TIME, and only when a pair node's persisted
	// tilt index is not one its ring has — a position.json written before the tilt became a
	// state, or by a build with a different lattice size. The node opens at the origin, and
	// this says which number was refused. At most one per node per load, and none at all for
	// a file written by this build at this size.
	BreadcrumbPairSeedUnknown
	// BreadcrumbPairLatticeAdopt: one per pair node per POINT-COUNT CHANGE, and only when
	// the index it was holding is not one the new lattice has — that node opens at the
	// origin instead, and this says which index was kept and what became of it. A change
	// that every node's index survives logs nothing.
	BreadcrumbPairLatticeAdopt
)

// BreadcrumbLabels is the single source of truth for the BreadcrumbLabel* enum's
// string names, indexed by the enum value — mirrored into TS by gen-node-defs for the
// .probe buffer-decoded breadcrumb log.
var BreadcrumbLabels = []string{
	"topology-loaded",
	"row-seed-count-mismatch",
	"pole-toggle-go",
	"window_clear",
	"window_open",
	"dwell_start",
	"abc-drag",
	"wire-send-buffer-full",
	"drag.commit",
	"wire-breadcrumbs-dropped",
	"chain-aim",
	"neighbor-center-recv",
	"neighbor-setc-recv",
	"bead-crud",
	"pair-seed-unknown",
	"pair-lattice-adopt",
}

// BreadcrumbLabelID resolves a breadcrumb's string name to its BreadcrumbLabel* index —
// the number a KindBreadcrumb RowEvent carries on the wire. It exists for the call sites
// that name their breadcrumb with the same string they pass to Trace.Breadcrumb, so the
// structured production emit and the test-sink line cannot name different things. ok is
// false for a name not in the table, which a caller should treat as "do not emit" rather
// than sending an id the decode side would resolve to the wrong label.
func BreadcrumbLabelID(name string) (uint8, bool) {
	for i, n := range BreadcrumbLabels {
		if n == name {
			return uint8(i), true
		}
	}
	return 0, false
}
