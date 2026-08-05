// Buffer/view_stream_frame.go — the VIEW stream's dedicated-frame packer (see
// Buffer/stream_fds.go's StreamKindView doc comment and
// memory/feedback_no_single_writer_bridge.md). Step C of retiring the old central
// accumulator (memory/feedback_no_single_writer_bridge.md): the
// gesture/stdin-reader goroutine (nodes/Wiring's MoveDispatch) already owns camera/overlay/
// scene-sphere/selection/hover state — it now WRITES this frame itself, directly, instead
// of routing that state through the old central accumulator's Trace-drain path.
//
// BuildViewStreamFrame produces the BYTE-IDENTICAL layout the old central accumulator's
// buildViewFrame used to produce (same SetCameraRow/SetOverlayRow/SetSceneRow column
// writers, same BuildEventsSection trailer) — built from plain values, mirroring
// BuildNodeStreamFrame/BuildEdgeStreamFrame's shape, so the emitting side (nodes/Wiring)
// needs only this one plain function, injected from main.go (which imports Buffer), to
// stay Buffer-independent itself.
package Buffer

import "encoding/binary"

// BuildViewStreamFrame packs the VIEW stream's own frame payload (no outer tag byte — the
// fd position already identifies the stream):
//
//	[tick:u32]
//	Camera  BufCameraStride bytes  (SAME SetCameraRow column writer buildViewFrame uses)
//	Overlay BufOverlayStride bytes (SAME SetOverlayRow column writer — see OverlayRow's
//	        doc comment)
//	Scene   BufSceneStride bytes   (SAME SetSceneRow column writer)
//	TABS section (BuildSceneTabsSection — see below)
//	EVENTS section (BuildEventsSection)
//
// The TABS section sits BEFORE the events trailer because BuildEventsSection's own contract
// is that it is always LAST in a frame (its text bytes carry no frame-level size
// bookkeeping, so nothing may follow them).
func BuildViewStreamFrame(tick uint32,
	camPX, camPY, camPZ, camR, camPosTheta, camPosPhi, camUpTheta, camUpPhi float32,
	overlay OverlayRow,
	sceneCX, sceneCY, sceneCZ, sceneRadius float32,
	tabNames []string, tabSelected uint16,
	events []StreamEvent,
) []byte {
	buf := make([]byte, BufViewFrameHeaderSize+BufCameraStride+BufOverlayStride+BufSceneStride)
	binary.LittleEndian.PutUint32(buf[0:], tick)
	off := BufViewFrameHeaderSize
	SetCameraRow(buf[off:], camPX, camPY, camPZ, camR, camPosTheta, camPosPhi, camUpTheta, camUpPhi)
	off += BufCameraStride
	SetOverlayRow(buf[off:], overlay)
	off += BufOverlayStride
	SetSceneRow(buf[off:], sceneCX, sceneCY, sceneCZ, sceneRadius)
	buf = append(buf, BuildSceneTabsSection(tabNames, tabSelected)...)
	return append(buf, BuildEventsSection(events)...)
}

// BuildSceneTabsSection packs the Go-owned scene tab strip:
//
//	[count:u16][selected:u16] then count × ( [nameLen:u16][name bytes] )
//
// A zero count (an untabbed anchor — see nodes/Wiring/scene_tabs.go's AnchorIsTabbed) still
// writes the two header fields, so the section's own width is never zero and the decoder
// never has to special-case "is there a tabs section at all". Names are the labels Go owns;
// TS renders them and never invents one.
func BuildSceneTabsSection(names []string, selected uint16) []byte {
	size := 4
	for _, n := range names {
		size += 2 + len(n)
	}
	buf := make([]byte, size)
	binary.LittleEndian.PutUint16(buf[0:], uint16(len(names)))
	binary.LittleEndian.PutUint16(buf[2:], selected)
	off := 4
	for _, n := range names {
		binary.LittleEndian.PutUint16(buf[off:], uint16(len(n)))
		off += 2
		off += copy(buf[off:], n)
	}
	return buf
}
