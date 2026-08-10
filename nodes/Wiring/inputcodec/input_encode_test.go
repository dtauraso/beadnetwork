// input_encode_test.go — the TEST-ONLY encoder for the editor→Go input stream.
//
// These exist only to build byte-exact fixtures for the decoder tests in this
// package (and for Wiring's own stdin-reader integration tests, which import this
// package for them). Go never encodes an input record in production: the production
// encoder is the TS one (tools/topology-vscode/src/schema/input-layout.ts), and
// Go only ever DECODES (input_codec.go). They live in a _test.go file so the
// production source reads as the decoder it is.

package inputcodec

import (
	"encoding/binary"
	"math"
)

type recWriter struct{ b []byte }

func (w *recWriter) u8(v byte)     { w.b = append(w.b, v) }
func (w *recWriter) i32(v int32)   { w.b = binary.LittleEndian.AppendUint32(w.b, uint32(v)) }
func (w *recWriter) f64(v float64) { w.b = binary.LittleEndian.AppendUint64(w.b, math.Float64bits(v)) }
func (w *recWriter) boolByte(v bool) {
	if v {
		w.u8(1)
	} else {
		w.u8(0)
	}
}

func enumIndex(list []string, s string) byte {
	for i, v := range list {
		if v == s {
			return byte(i)
		}
	}
	return 0
}

// EncodeControl builds a payload-less control record (save).
func EncodeControl(kind byte) []byte { return []byte{kind} }

// EncodeOverlaysToggle builds an overlays TOGGLE record (test helper).
func EncodeOverlaysToggle(flag string) []byte {
	w := &recWriter{}
	w.u8(InKindEditUpdate)
	w.u8(enumIndex(InUpdateKinds, "overlays"))
	w.u8(InOverlayAttrToggle)
	w.u8(enumIndex(InOverlayFlags, flag))
	return w.b
}

// EncodeDistanceGroupAdjust builds a distanceGroup LENGTH record (test helper):
// [22][entityKind=distanceGroup][attr=length][u8 groupIndex][u8 dirUp].
func EncodeDistanceGroupAdjust(groupIdx int, dirUp bool) []byte {
	w := &recWriter{}
	w.u8(InKindEditUpdate)
	w.u8(enumIndex(InUpdateKinds, "distanceGroup"))
	w.u8(InDistanceGroupAttrLength)
	w.u8(byte(groupIdx))
	if dirUp {
		w.u8(1)
	} else {
		w.u8(0)
	}
	return w.b
}

// EncodeSceneLatticePoints builds a scene LATTICE-POINTS record (test helper):
// [22][entityKind=scene][attr=latticePoints][u8 points].
func EncodeSceneLatticePoints(points int) []byte {
	w := &recWriter{}
	w.u8(InKindEditUpdate)
	w.u8(enumIndex(InUpdateKinds, "scene"))
	w.u8(InSceneAttrLatticePoints)
	w.u8(byte(points))
	return w.b
}

// EncodeRawInput builds a raw-input record from a RawInputMsg (test helper).
func EncodeRawInput(ev RawInputMsg) []byte {
	w := &recWriter{}
	w.u8(InKindRawInput)
	w.u8(enumIndex(InEventKinds, ev.Kind))
	w.f64(ev.X)
	w.f64(ev.Y)
	w.f64(ev.RectLeft)
	w.f64(ev.RectTop)
	w.f64(ev.RectWidth)
	w.f64(ev.RectHeight)
	w.i32(int32(ev.Button))
	w.boolByte(ev.Ctrl)
	w.boolByte(ev.Shift)
	w.boolByte(ev.Alt)
	w.boolByte(ev.Meta)
	w.f64(ev.DeltaX)
	w.f64(ev.DeltaY)
	w.f64(ev.Fov)
	w.u8(enumIndex(InHitKinds, ev.Hit.Kind))
	w.boolByte(ev.Hit.IsInput)
	w.i32(int32(ev.Hit.NodeRow))
	w.i32(int32(ev.Hit.PortRow))
	w.i32(int32(ev.Hit.EdgeRow))
	return w.b
}

// FrameRecord wraps a record body with the [len:u32-LE] transport frame.
func FrameRecord(rec []byte) []byte {
	return append(binary.LittleEndian.AppendUint32(nil, uint32(len(rec))), rec...)
}
