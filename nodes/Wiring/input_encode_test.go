// input_encode_test.go — the TEST-ONLY encoder for the editor→Go input stream.
//
// These exist only to build byte-exact fixtures for the decoder tests in this
// package. Go never encodes an input record in production: the production
// encoder is the TS one (tools/topology-vscode/src/schema/input-layout.ts), and
// Go only ever DECODES (input_codec.go). They live in a _test.go file so the
// production source reads as the decoder it is.

package Wiring

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

// encodeControl builds a payload-less control record (save).
func encodeControl(kind byte) []byte { return []byte{kind} }

// encodeOverlaysToggle builds an overlays TOGGLE record (test helper).
func encodeOverlaysToggle(flag string) []byte {
	w := &recWriter{}
	w.u8(inKindEditUpdate)
	w.u8(enumIndex(inUpdateKinds, "overlays"))
	w.u8(inOverlayAttrToggle)
	w.u8(enumIndex(inOverlayFlags, flag))
	return w.b
}

// encodeDistanceGroupAdjust builds a distanceGroup LENGTH record (test helper):
// [22][entityKind=distanceGroup][attr=length][u8 groupIndex][u8 dirUp].
func encodeDistanceGroupAdjust(groupIdx int, dirUp bool) []byte {
	w := &recWriter{}
	w.u8(inKindEditUpdate)
	w.u8(enumIndex(inUpdateKinds, "distanceGroup"))
	w.u8(inDistanceGroupAttrLength)
	w.u8(byte(groupIdx))
	if dirUp {
		w.u8(1)
	} else {
		w.u8(0)
	}
	return w.b
}

// encodeSceneLatticePoints builds a scene LATTICE-POINTS record (test helper):
// [22][entityKind=scene][attr=latticePoints][u8 points].
func encodeSceneLatticePoints(points int) []byte {
	w := &recWriter{}
	w.u8(inKindEditUpdate)
	w.u8(enumIndex(inUpdateKinds, "scene"))
	w.u8(inSceneAttrLatticePoints)
	w.u8(byte(points))
	return w.b
}

// encodeRawInput builds a raw-input record from a rawInputMsg (test helper).
func encodeRawInput(ev rawInputMsg) []byte {
	w := &recWriter{}
	w.u8(inKindRawInput)
	w.u8(enumIndex(inEventKinds, ev.Kind))
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
	w.u8(enumIndex(inHitKinds, ev.Hit.Kind))
	w.boolByte(ev.Hit.IsInput)
	w.i32(int32(ev.Hit.NodeRow))
	w.i32(int32(ev.Hit.PortRow))
	w.i32(int32(ev.Hit.EdgeRow))
	return w.b
}

// frameRecord wraps a record body with the [len:u32-LE] transport frame.
func frameRecord(rec []byte) []byte {
	return append(binary.LittleEndian.AppendUint32(nil, uint32(len(rec))), rec...)
}
