package streamframe

import (
	"encoding/binary"

	B "github.com/dtauraso/wirefold/Buffer"
	T "github.com/dtauraso/wirefold/Trace"
)

type StreamEvent struct {
	Kind                                                             uint8
	NodeRow, PortRow, TargetRow, TargetPortRow, EdgeRow, Slot, Value int32
	Bead                                                             uint32

	BeadSteps  float32
	X, Y, Z, F float32

	Label uint8
	Debug uint8
	Text  string
}

var kindIDByName = buildKindIDMap()

func buildKindIDMap() map[string]uint8 {
	m := make(map[string]uint8, len(T.TraceEventKinds))
	for i, k := range T.TraceEventKinds {
		m[k] = uint8(i)
	}
	return m
}

func KindID(kind string) uint8 {
	return kindIDByName[kind]
}

func BuildEventsSection(events []StreamEvent) []byte {
	textBytes := make([][]byte, len(events))
	textLen := 0
	for i, e := range events {
		textBytes[i] = []byte(e.Text)
		textLen += len(textBytes[i])
	}
	buf := make([]byte, 4+len(events)*B.BufEventStride+textLen)
	binary.LittleEndian.PutUint32(buf[0:], uint32(len(events)))
	rows := buf[4 : 4+len(events)*B.BufEventStride]
	textOff := uint32(0)
	off := 4 + len(events)*B.BufEventStride
	for i, e := range events {
		tb := textBytes[i]
		B.SetEventRow(rows, i,
			e.Kind, e.NodeRow, e.PortRow, e.TargetRow, e.TargetPortRow, e.EdgeRow,
			e.Slot, e.Value, e.Bead, e.BeadSteps, e.X, e.Y, e.Z, e.F,
			e.Label, e.Debug, textOff, uint32(len(tb)))
		copy(buf[off:], tb)
		off += len(tb)
		textOff += uint32(len(tb))
	}
	return buf
}
