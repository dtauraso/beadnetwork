// move_dispatch_construct.go — builds a MoveDispatch from load-time geometry: one
// nodeMover per node, one edgeMover per edge, and the dedicated directed channels wiring
// adjacent movers together (see move_dispatch.go's doc comment for the model this
// reproduces per-goroutine). This is the one-time, single-threaded setup step; the
// MoveDispatch struct itself lives in move_dispatch.go and its public delegator API lives
// in move_dispatch_api.go. The phases NewMoveDispatch calls in order are grouped by
// concern into sibling files: move_dispatch_seeds.go (order resolution, buffer row seeds,
// and the row-identity tables built from them), move_dispatch_movers.go (constructing the
// node/edge movers and wiring the channels/pairs/ids between them), and
// move_dispatch_ui.go (UI-state defaults and the closures EmitViewFrame needs). This file
// keeps NewMoveDispatch itself, the sole constructor of the &MoveDispatch{...} struct
// literal.

package dispatch

import (
	T "github.com/dtauraso/wirefold/Trace"
	"github.com/dtauraso/wirefold/nodes/Wiring/inputcodec"
	"github.com/dtauraso/wirefold/nodes/Wiring/moverreg"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
	"github.com/dtauraso/wirefold/nodes/wire/clock"
)

// NewMoveDispatch builds the registry from per-node geometry and per-edge endpoints.
// Exported (this task) so nodes/Wiring/build's buildMoveDispatch can call it: it
// constructs the MoveDispatch struct literal itself, which is why this constructor stays
// in package dispatch even though its caller moved out (every MoveDispatch field is
// exported — §35, docs/planning/movedispatch-decomposition.md removed the last unexported
// field, tapToInstall — but the struct LITERAL still names the type, so the constructor
// stays where the type is declared).
// It creates one nodeMover per node and one edgeMover per edge, registering each under
// its key (node id / edge id) in md.MR.NodeGeoms()/md.MR.EdgeMovers(), and wires the dedicated
// directed channels between adjacent movers. Outs and dest wires are bound later by Bind once node
// construction has populated them. nodeOrder/edgeOrder are the
// SPEC order (deterministic directory-sorted order, not map iteration order) used to
// build md.GS.NodeSeeds/EdgeSeeds for buffer row seeding.
//
// speedSinks, when non-nil, is the loader's build-wide accumulator
// (buildCtx.speedSinks): each nodeMover AND each edgeMover created below gets its own
// fresh buffered-1 speed channel (per-goroutine-clock.md "Delivery" — every
// clock-owning goroutine must not be left behind), and that channel's SEND end is
// appended here.
// nil in test call sites that construct a MoveDispatch directly with no
// loader — those edgeMovers then simply have no speed channel to poll.
// rowCount is the buffer's node-row space (topoSpec.RowCount — the largest node id found,
// not the node count): rows 0..rowCount-1, ROW ID = NODE ID - 1. 0 (test call sites that
// don't pass one) falls back to the number of resolved seeds, i.e. no gaps.
func NewMoveDispatch(geoms map[string]nodegeom.NodeGeom, edgeEndpoints map[string]inputcodec.EdgeEndpoints, tr *T.Trace, nodeOrder, edgeOrder []string, clk clock.Clock, speedSinks *[]chan float64, rowCount int) (*MoveDispatch, error) {
	nodeOrder, edgeOrder = resolveSeedOrders(geoms, edgeEndpoints, nodeOrder, edgeOrder)

	md := &MoveDispatch{
		TR: tr,
	}
	md.MR = moverreg.New()
	initMoveDispatchUIDefaults(md)

	if err := md.buildGeomSeeds(geoms, edgeEndpoints, nodeOrder, edgeOrder); err != nil {
		return nil, err
	}
	md.buildNodeMovers(geoms, tr, clk)
	md.wireMutualPairs(edgeEndpoints)
	md.buildEdgeMovers(edgeEndpoints, geoms, tr, clk, speedSinks)
	md.seedPartnerCenters()
	md.wireNodeEdgeIDs()
	md.buildRowTables(rowCount)
	md.bindUIClosures()

	return md, nil
}
