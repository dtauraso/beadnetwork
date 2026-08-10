// input_fingerprint.go — the LAYOUT PIN for the editor→Go input stream.
//
// One job: say what the wire vocabulary IS. The record kind bytes, the update-attr
// indices, and the five enum orderings all derive from a single string constant here, and
// nothing in this file reads a byte off the wire — that is rec_reader.go's job, and
// turning bytes into a StdinMsg is input_codec.go's.
//
// This is also the file tools/gen-node-defs reads (via go/ast) to GENERATE the TS side's
// input-layout-gen.ts, so the fingerprint below is the single source for both languages.

package inputcodec

import "strings"

// InputLayoutFingerprint pins the binary input-record layout. It is the SINGLE source: the
// TS-side INPUT_LAYOUT_FINGERPRINT in input-layout-gen.ts is GENERATED from this string
// (tools/gen-node-defs/input_layout.go), so it always matches by construction — regenerate
// (npm run gen:node-defs) after changing any record kind, field, or enum ordering.
//
// No leading version token: it was a hand-bumped marker nothing parsed (parseFPList reads
// the kinds=/eventKinds=/… tokens, never a version), and record FIELDS aren't listed here
// anyway — so a field-layout change (e.g. removing a hit field) wouldn't move this string.
// The real Go↔TS byte-alignment guard is TestInputFixtureCrossLanguage, which drives a
// record through both encoders and diffs the fields; do not re-add a version pretending to
// be a checked invariant.
//
// INPUT_LAYOUT_FINGERPRINT: kinds=save:4,raw-input:10,edit-update:22 eventKinds=pointerdown,pointermove,pointerup,wheel,home hitKinds=port,handhold,node,edge,torus,empty updateKinds=overlays,clock,distanceGroup,scene,tiltVector updateAttrs=toggle,speed,length,selected,theta,phi,reset,start,latticePoints,create,delete overlayFlags=tori,scenePoles,nodePoles,selSpherePoles,handholds,labelsGlobal,overlays,nodeBody,nodeRing,ringPick,selectionRing,hoverRing,reachSphere
const InputLayoutFingerprint = "kinds=save:4,raw-input:10,edit-update:22 eventKinds=pointerdown,pointermove,pointerup,wheel,home hitKinds=port,handhold,node,edge,torus,empty updateKinds=overlays,clock,distanceGroup,scene,tiltVector updateAttrs=toggle,speed,length,selected,theta,phi,reset,start,latticePoints,create,delete overlayFlags=tori,scenePoles,nodePoles,selSpherePoles,handholds,labelsGlobal,overlays,nodeBody,nodeRing,ringPick,selectionRing,hoverRing,reachSphere"

// Record kind bytes (first byte of every record).
const (
	// Kinds 1 (resume) and 2 (pause) removed — the play/pause clock gate was deleted
	// end-to-end. Intentional gaps, per house style (never renumber a live wire value).
	// Kind 3 (InKindResend) removed — intentional gap, see comment above.
	InKindSave = 4 // save  — Go persists its OWN scene state (bare command)
	// Kind 5 (InKindFadeToggle) removed — the fade feature was deleted end-to-end.
	InKindRawInput = 10 // raw pointer/wheel/home event
	// Kind 20 (InKindEditCreate) removed — edge creation via edit op was deleted
	// end-to-end. Intentional gap, per house style (never renumber a live wire value).
	// Kind 21 (InKindEditDelete) removed — same removal as above.
	InKindEditUpdate = 22 // edit op=update (entity byte + attr byte + numeric payload)
)

// Update attr indices (positional in IN_UPDATE_ATTRS; the order is pinned across both
// languages by the updateAttrs= token in the fingerprint).
const (
	InOverlayAttrToggle       = 0 // overlays: flip one flag
	InClockAttrSpeed          = 1 // clock: set the playback-speed multiplier
	InDistanceGroupAttrLength = 2 // distanceGroup: set the group's target pair length
	InSceneAttrSelected       = 3 // scene: select a tab from the Go-owned scene tab strip
	InTiltVectorAttrTheta     = 4 // tiltVector: adjust the vector's θ index by one click
	// attr 5 (φ) is a GAP — the tilt vector is θ-only end to end now
	// (task/drop-tilt-vector-phi); never renumber the survivors.
	InTiltVectorAttrReset    = 6 // tiltVector: return the index to 0 (the RESET button)
	InTiltVectorAttrStart    = 7 // tiltVector: begin the vector exchange from the current angles (the START TILT button)
	InSceneAttrLatticePoints = 8 // scene: set the pair lattice's point count
	// scene: CREATE a node at a dropped world point, and DELETE the node on a buffer row.
	// ATTRIBUTES, not the create/delete OPS that were removed end-to-end (their kind bytes
	// 20/21 stay gaps and are not reused): the edit surface still has exactly one op, and a
	// new capability is a new entity kind or attribute (CLAUDE.md's bridge-surface rule).
	InSceneAttrCreate = 9  // scene: create a node of a kind, at a world point
	InSceneAttrDelete = 10 // scene: delete the node on a buffer row, and every edge touching it
)

// Enum orderings (u8 index → string), shared with input-layout-gen.ts. All five orderings
// (eventKinds/hitKinds/updateKinds/updateAttrs/overlayFlags) are DERIVED from their token in
// the fingerprint, so a Go-side ordering CANNOT drift from the pinned layout: there is no
// second array to reorder. The chain that keeps both languages in lockstep, per enum:
//
//	Go fingerprint (the source)  →[parseFPList, here]→  Go array
//	   Go fingerprint  →[tools/gen-node-defs/input_layout.go parses the same string]→  TS array
//
// Every link is DERIVED from the one Go fingerprint string — there is no second
// hand-authored copy on either side to fall out of sync. These orderings are WIRE INDICES (a u8
// index is all that crosses the bridge), so an unchecked reorder is a silent
// mis-dispatch — a raycast hit on a node decoding as an edge — with nothing to fail.
var (
	InEventKinds   = parseFPList(InputLayoutFingerprint, "eventKinds=")
	InHitKinds     = parseFPList(InputLayoutFingerprint, "hitKinds=")
	InUpdateKinds  = parseFPList(InputLayoutFingerprint, "updateKinds=")
	InUpdateAttrs  = parseFPList(InputLayoutFingerprint, "updateAttrs=")
	InOverlayFlags = parseFPList(InputLayoutFingerprint, "overlayFlags=")
)

// init fails the process at STARTUP if any enum token is missing from the fingerprint.
// Without this a typo'd/renamed token yields a nil list, enumAt returns "", and every
// record carrying that enum is rejected at runtime — a live input bridge that silently
// does nothing. A malformed layout is a build/boot error, not a quiet degradation
// (mirrors gen-node-defs: a malformed wire prop tag is an error, not a silent skip).
func init() {
	for _, e := range []struct {
		marker string
		list   []string
	}{
		{"eventKinds=", InEventKinds},
		{"hitKinds=", InHitKinds},
		{"updateKinds=", InUpdateKinds},
		{"updateAttrs=", InUpdateAttrs},
		{"overlayFlags=", InOverlayFlags},
	} {
		if len(e.list) == 0 {
			panic("input_codec: INPUT_LAYOUT_FINGERPRINT is missing the " + e.marker + " token — the wire enum orderings derive from it")
		}
	}
}

// parseFPList extracts one space-delimited, comma-separated enum token from the
// fingerprint (e.g. marker "hitKinds=" → ["port","handhold",...]). Returns nil if the
// marker is absent; init() above turns that into a startup panic.
func parseFPList(fp, marker string) []string {
	i := strings.Index(fp, marker)
	if i < 0 {
		return nil
	}
	rest := fp[i+len(marker):]
	if sp := strings.IndexByte(rest, ' '); sp >= 0 {
		rest = rest[:sp]
	}
	return strings.Split(rest, ",")
}
