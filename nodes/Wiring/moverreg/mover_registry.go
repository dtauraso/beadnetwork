// mover_registry.go — the nodeMover/edgeMover directory owner, lifted out of package
// dispatch into its own package (docs/planning/movedispatch-decomposition.md §25),
// mirroring the §17 edgemover / §20 nodeactor package moves: same "still one goroutine per
// node/edge, still the same channels, only the package boundary moved" shape. MoverRegistry
// owns nodeGeoms/nodeMovers/edgeMovers/edgeOut/selfDriveClaimed/centerMirror. Every
// external touch that used to reach a bare field (md.mr.nodeGeoms[...],
// md.mr.edgeMovers[...], ...) now goes through NodeGeoms()/EdgeMovers() (both return the
// LIVE map — a map is a reference type, so an external `mr.NodeGeoms()[id] = ng` still
// writes MoverRegistry's own map, exactly as the old bare field write did) or one of the
// construction-time wiring methods below (ClaimSelfDrive, SeedCenter) that fold what used
// to be a multi-statement external write into one call, the same shape §19's
// node_geometry_wire.go used for nodeGeometry's own construction-time writes.
//
// This file holds the struct + directory accessors + construction-time writes only.
// Bind/Start/FinalizeActors (wiring and actor launch) live in mover_registry_wire.go;
// the read-only lookups (CenterOfNode, SendMove, NodeBodyRadius, ...) live in
// mover_registry_query.go; the pure LinkRefusal decision lives in
// mover_registry_linkrefusal.go — one same-package split by concern, no API change (§26
// pattern, mirroring the Trace/Trace.go split).
package moverreg

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/edgemover"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor"
	wire "github.com/dtauraso/wirefold/nodes/wire"
)

// InboxDepth is the declared capacity of every per-mover movemsg.Msg inbox: an edgeMover's
// extIn/srcIn/dstIn (nodes/Wiring/edgemover), a nodeactor.NodeGeometry's extIn (nodeactor's
// own copy, nodeactor/consts.go), and each directed neighborIn/tiltEditIn/latticeIn built
// in package dispatch (move_dispatch_construct.go, build_nodes.go). Exported (was
// dispatch's own unexported moverInboxDepth) because build_nodes.go, in package dispatch,
// still constructs tiltEdit/lattice inbox channels at the same depth and has no other route
// to this number now that the type that names it best lives in this package — same
// "duplicate a small owned constant across an actor package boundary" call as edgemover's
// own InboxDepth (§17) and nodeactor's inboxDepth duplicate (§20), except here there was
// already exactly one definition, so it is exported rather than duplicated.
//
// WHY THIS NUMBER, honestly: it is a chosen depth, NOT derived. What IS derivable from the
// topology is the COUNT of these channels (one triple per edge, one extIn per node, two
// directed neighborIn per edge — 10/9/20 for the shipped graph), and the loader already
// fixes that by construction. The DEPTH is a queue for a burst of move messages during a
// drag, so it is bounded by gesture rate, not by anything in the spec files
// (docs/planning/visual-editor/session-log.md classifies it DYNAMIC for exactly this
// reason). 8 is "a few frames of drag messages"; it has held in practice and no
// measurement yet contradicts it.
const InboxDepth = 8

// MoverRegistry is the pure registry that owns every mover and wires their dedicated
// channels together — there is no shared dispatch map; nodeMovers/edgeMovers themselves
// are the directories a mover's resolveDest closure and the external-entry helpers below
// look up. It also retains the per-edge source Outs so in-package callers can read an
// edge's loaded geometry (edgeOutFor) without going through a central coordinator.
type MoverRegistry struct {
	// nodeGeoms is the UNIVERSAL per-node directory — every node's own
	// *nodeactor.NodeGeometry, ring and pair alike. This is what routing (resolveDest,
	// SendMove, CenterOfNode, NodeKind, drag/commit) looks up: a message addressed to a
	// node must arrive and be handled regardless of which goroutine drives that node's
	// geometry.
	nodeGeoms map[string]*nodeactor.NodeGeometry
	// nodeMovers is the RING-ONLY actor directory: one entry per node whose OWN kind did
	// NOT claim BuildArgs.ClaimSelfDrive, populated by FinalizeActors AFTER buildNodes has
	// run (so every ClaimSelfDrive call has already happened). Used ONLY by Start() to
	// launch a dedicated goroutine per ring node — a PAIR node (PairNode) has NO entry
	// here at all, by construction, not by a flag that says "launch nothing for me".
	nodeMovers map[string]*nodeactor.NodeMover
	// selfDriveClaimed holds, for each node id whose OWN kind claimed
	// BuildArgs.ClaimSelfDrive at build time (PairNode — the pair scene), true. Written
	// ONCE per entry by ClaimSelfDrive, on the single-threaded build path, before
	// FinalizeActors runs and before any goroutine exists. FinalizeActors reads it to
	// decide which nodes get a nodeMover actor at all — an id present here gets NONE.
	selfDriveClaimed map[string]bool
	edgeMovers       map[string]*edgemover.EdgeMover
	// edgeOut: edgeId → source *Out, for read-only access by tests/verifiers.
	edgeOut map[string]*wire.Out
	// centerMirror is the DISPATCH goroutine's OWN mirror of every node's last-known
	// world center, kept current by messages from each node's own goroutine.
	// Seeded once at construction (single-threaded setup, from each node's load-time
	// geom, via SeedCenter) so the first framing read has every center before any push
	// arrives, then kept current by drainCenterMirror pulling each nodeGeometry's own
	// centerOut channel (via PollCenter). Written and read ONLY from the dispatch/gesture
	// goroutine (CenterOfNode is, after the quantize call sites moved to each node's own
	// partnerCenters map, called only from that goroutine) — no lock.
	centerMirror map[string]vec3
}

// New returns an empty, ready-to-wire MoverRegistry — every map initialized, nothing
// registered yet. Replaces the 4-line map-init block that used to sit at the top of
// package dispatch's newMoveDispatch (construction-time write consolidation, same move
// §19 did for nodeGeometry's own construction-time writes).
func New() MoverRegistry {
	return MoverRegistry{
		nodeGeoms:    map[string]*nodeactor.NodeGeometry{},
		edgeMovers:   map[string]*edgemover.EdgeMover{},
		edgeOut:      map[string]*wire.Out{},
		centerMirror: map[string]vec3{},
	}
}

// NodeGeoms returns the LIVE per-node directory (every node's own *nodeactor.NodeGeometry).
// A map is a reference type, so an external `mr.NodeGeoms()[id] = ng` or
// `for id, ng := range mr.NodeGeoms()` reaches the SAME map MoverRegistry itself reads and
// writes — this is the mechanical replacement for every former bare `md.mr.nodeGeoms`
// touch, not a copy.
func (mr *MoverRegistry) NodeGeoms() map[string]*nodeactor.NodeGeometry {
	return mr.nodeGeoms
}

// EdgeMovers returns the LIVE per-edge directory, same reference-type reasoning as
// NodeGeoms.
func (mr *MoverRegistry) EdgeMovers() map[string]*edgemover.EdgeMover {
	return mr.edgeMovers
}

// ClaimSelfDrive marks node id as self-driven (its OWN kind's goroutine drives its
// geometry — PairNode — rather than a separate nodeactor.NodeMover). Folds what used to be
// package dispatch's 3-statement external write (nil-check, lazy-init, set) into one call —
// the construction-time write consolidation §19 did for nodeGeometry, applied here to the
// one selfDriveClaimed write this package's own external caller (build_nodes.go's
// ClaimSelfDrive path) makes.
func (mr *MoverRegistry) ClaimSelfDrive(id string) {
	if mr.selfDriveClaimed == nil {
		mr.selfDriveClaimed = map[string]bool{}
	}
	mr.selfDriveClaimed[id] = true
}

// SeedCenter records id's initial world center in the center mirror, at construction
// (before any node goroutine has pushed a live update) — package dispatch's one
// construction-time centerMirror write.
func (mr *MoverRegistry) SeedCenter(id string, c vec3) {
	mr.centerMirror[id] = c
}

// Bind, Start, FinalizeActors: see mover_registry_wire.go.
// drainCenterMirror, CenterOfNode, SendMove, EnqueueFor, nodeKind, NodeBodyRadius,
// HasNodeMover, NodeSelfDriven, NodeQuantOffset, NearestNodeTo: see mover_registry_query.go.
// LinkRefusal, linkRefusalFor, firstPortOfDir: see mover_registry_linkrefusal.go.
