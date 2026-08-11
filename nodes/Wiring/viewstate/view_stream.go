// view_stream.go — the VIEW stream's write side, lifted from Wiring's view_stream.go into
// this package (docs/planning/gesture-actor.md; the earlier "uiState/overlayState declined"
// probes — docs/planning/movedispatch-decomposition.md sections 5/6b — were blocked
// specifically by THIS file's direct field access to md.ui.ov/md.ui.vp/md.ui.sceneSphere;
// moving it here alongside UIState/OverlayState is what unblocks the lift).
//
// Camera/overlay/scene-sphere/selection/hover state lives on UIState (ui.VP/ui.OV/
// ui.SceneSphere), mutated only by the gesture/stdin-reader goroutine (Wiring's
// RunStdinReader single dispatch loop — UNCHANGED by this lift, see this package's own
// header note below). This file is the WRITE side: pack that state into the VIEW stream's
// own frame and write it to the dedicated view fd whenever it changes.
//
// VIEW IS EVENT-DRIVEN, AND IT IS THE ODD ONE OUT — see Wiring's original view_stream.go
// history for the full "no clock/tick-driven emit" reasoning, carried over unchanged.
//
// OWNERSHIP UNCHANGED: this file only ADDS a package boundary around the exact same state
// and the exact same write. The goroutine that calls EmitViewFrame/EmitBreadcrumb/
// SetViewStream is still, and only ever, Wiring's RunStdinReader dispatch goroutine — this
// package holds no goroutine of its own. runtopology/topology_run.go's startup-seed-before-
// actor-exists ordering (emitStartupBreadcrumbs/loadSceneState at lines 63/65, before
// startStdinReader creates the actor at line 80) is untouched: those calls still resolve
// through MoveDispatch (md.UI.EmitBreadcrumb / md.UI.SetViewStream, exported field), still on
// the same startup goroutine, still strictly before any other goroutine exists.
package viewstate

import (
	"encoding/binary"
	"io"

	wire "github.com/dtauraso/wirefold/nodes/wire"

	T "github.com/dtauraso/wirefold/Trace"
)

// ViewFrameBuilder packs the VIEW stream's own frame payload from plain values (mirrors
// nodeMover/edgeMover's buildFrame closures — injected from main.go, which imports Buffer,
// so this package stays Buffer-independent).
type ViewOverlayFlags struct {
	SceneTori, ScenePoles, NodePoles, SelSpherePoles, Handholds, LabelsGlobal, OverlaysVis uint8
	NodeBody, NodeRing, RingPick, SelectionRing, HoverRing, ReachSphere                    uint8
}

// ViewSceneState carries the VIEW frame's per-scene scalars that are neither camera, nor
// overlay flags, nor a panel readout.
type ViewSceneState struct {
	EditRefused   uint32
	SceneEditable uint8
	SceneKinds    uint32
}

type ViewFrameBuilder func(tick uint32,
	camPX, camPY, camPZ, camR, camPosTheta, camPosPhi, camUpTheta, camUpPhi float32,
	flags ViewOverlayFlags,
	dragNodeRow int32,
	scene ViewSceneState,
	groupLenTime, groupLenInput, groupLenGate float32,
	speed float32,
	sceneCX, sceneCY, sceneCZ, sceneRadius float32,
	events []wire.RowEvent,
) []byte

// SetViewStream installs the VIEW stream's write side: out is the dedicated view fd (nil =
// no dedicated stream, so EmitViewFrame early-returns as a no-op) and buildFrame packs this
// goroutine's own frame bytes (Buffer.BuildViewStreamFrame, injected from main.go). Call
// once at startup, before any gesture/edit reaches RunStdinReader.
func (ui *UIState) SetViewStream(out io.Writer, buildFrame ViewFrameBuilder) {
	// "view" is a singleton stream — a second SetViewStream call is rejected the same way a
	// second setNodeStreams/setEdgeStreams claim on one row is (Wiring's stream_claim.go).
	ui.viewOut = newViewClaimedStream(&ui.viewClaimed, out)
	ui.ViewBuildFrame = buildFrame
}

// EmitBreadcrumb writes ev as a structured Breadcrumb event on the VIEW stream (Kind/Debug
// are forced regardless of what the caller passed). No-op (via EmitViewFrame) when the VIEW
// stream isn't wired.
func (ui *UIState) EmitBreadcrumb(ev wire.RowEvent) {
	ev.Kind = T.KindBreadcrumb
	ev.Debug = 1
	ui.EmitViewFrame([]wire.RowEvent{ev})
}

// EmitViewFrame packs and writes the current camera/overlay/scene-sphere state as this
// goroutine's own VIEW frame, if the dedicated stream is active (nil ViewBuildFrame — no
// WIREFOLD_STREAM_FDS "view" entry — is the required no-op fallback). events carries
// whatever this call's OWN state change should log, resolved to buffer rows by the caller.
func (ui *UIState) EmitViewFrame(events []wire.RowEvent) {
	if ui.ViewBuildFrame == nil {
		return
	}
	ui.viewTick++
	v := ui.VP.Viewpoint
	sc := ui.SceneSphere
	// dragNodeRow is derived from LastDraggedNode, NOT the live Gest.DragNode — see
	// LastDraggedNode's own doc comment (ui_state.go).
	dragNodeRow := int32(-1)
	if ui.LastDraggedNode != "" && ui.NodeRowFor != nil {
		if r, ok := ui.NodeRowFor(ui.LastDraggedNode); ok {
			dragNodeRow = r
		}
	}
	// The "distance home button" panel's 3 group max-pair-lengths, recomputed fresh from
	// live node centers on every VIEW-frame emit via the bound closure (see
	// DistanceGroupLensFn's own doc comment — Wiring's DistanceGroupLens itself stays in
	// Wiring, it reads *moverRegistry, an unexported Wiring type this package cannot name).
	var groupLenTime, groupLenInput, groupLenGate float32
	if ui.DistanceGroupLensFn != nil {
		groupLenTime, groupLenInput, groupLenGate = ui.DistanceGroupLensFn()
	}
	frame := ui.ViewBuildFrame(ui.viewTick,
		float32(v.Pivot.X), float32(v.Pivot.Y), float32(v.Pivot.Z), float32(v.R),
		float32(v.Pos.Theta), float32(v.Pos.Phi), float32(v.Up.Theta), float32(v.Up.Phi),
		ViewOverlayFlags{
			SceneTori:      boolU8(ui.OV.SceneToriVisible),
			ScenePoles:     boolU8(ui.OV.ScenePolesVisible),
			NodePoles:      boolU8(ui.OV.NodePolesVisible),
			SelSpherePoles: boolU8(ui.OV.SelSpherePolesVisible),
			Handholds:      boolU8(ui.OV.HandholdsVisible),
			LabelsGlobal:   boolU8(ui.OV.LabelsGlobalVisible),
			OverlaysVis:    boolU8(ui.OV.OverlaysVisible),
			NodeBody:       boolU8(ui.OV.NodeBodyVisible),
			NodeRing:       boolU8(ui.OV.NodeRingVisible),
			RingPick:       boolU8(ui.OV.RingPickVisible),
			SelectionRing:  boolU8(ui.OV.SelectionRingVisible),
			HoverRing:      boolU8(ui.OV.HoverRingVisible),
			ReachSphere:    boolU8(ui.OV.ReachSphereVisible),
		},
		dragNodeRow,
		ViewSceneState{
			EditRefused:   ui.EditRefused,
			SceneEditable: boolU8(ui.SceneEditable),
			SceneKinds:    ui.SceneKinds,
		},
		groupLenTime, groupLenInput, groupLenGate,
		float32(ui.Speed),
		float32(sc.Center.X), float32(sc.Center.Y), float32(sc.Center.Z), float32(sc.Radius),
		events,
	)
	if !ui.viewOut.Ok() {
		return
	}
	var hdr [4]byte
	binary.LittleEndian.PutUint32(hdr[:], uint32(len(frame)))
	// Fire-and-forget, same reasoning as every other stream's frame write in this codebase:
	// no delivery guarantee on this channel, errors ignored.
	_, _ = ui.viewOut.Write(hdr[:])
	_, _ = ui.viewOut.Write(frame)
}

// boolU8 converts a bool to a 0/1 byte for the buffer's flag columns.
func boolU8(b bool) uint8 {
	if b {
		return 1
	}
	return 0
}
