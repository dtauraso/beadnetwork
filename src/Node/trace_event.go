package Node

import (
	"encoding/binary"
	"math"
)

type RowEvent struct {
	Kind                                                             string
	NodeRow, PortRow, TargetRow, TargetPortRow, EdgeRow, Slot, Value int32
	Bead                                                             uint64

	BeadSteps float64
	X, Y, Z   float64

	Label string
	Debug uint8
	Text  string
}

const (
	KindRecv       = "recv"
	KindFire       = "fire"
	KindSend       = "send"
	KindArrive     = "arrive"
	KindBreadcrumb = "breadcrumb"
)

var TraceEventKinds = []string{KindRecv, KindFire, KindSend, KindArrive, KindBreadcrumb}

const TraceRecordFixedSize = 82

func kindID(kind string) uint8 {
	for i, k := range TraceEventKinds {
		if k == kind {
			return uint8(i + 1)
		}
	}
	return 0
}

func AppendRecord(buf []byte, e RowEvent, nowMs int64) []byte {
	var rec [TraceRecordFixedSize]byte

	rec[0] = kindID(e.Kind)
	rec[1] = e.Debug

	label := e.Label
	if len(label) > math.MaxUint8 {
		label = label[:math.MaxUint8]
	}
	rec[2] = uint8(len(label))

	putI32(rec[4:], e.NodeRow)
	putI32(rec[8:], e.PortRow)
	putI32(rec[12:], e.TargetRow)
	putI32(rec[16:], e.TargetPortRow)
	putI32(rec[20:], e.EdgeRow)
	putI32(rec[24:], e.Slot)
	putI32(rec[28:], e.Value)

	binary.LittleEndian.PutUint64(rec[32:], uint64(nowMs))
	binary.LittleEndian.PutUint64(rec[40:], e.Bead)

	putF64(rec[48:], e.BeadSteps)
	putF64(rec[56:], e.X)
	putF64(rec[64:], e.Y)
	putF64(rec[72:], e.Z)

	text := e.Text
	if len(text) > math.MaxUint16 {
		text = text[:math.MaxUint16]
	}
	binary.LittleEndian.PutUint16(rec[80:], uint16(len(text)))

	buf = append(buf, rec[:]...)
	buf = append(buf, label...)
	return append(buf, text...)
}

func putI32(b []byte, v int32) { binary.LittleEndian.PutUint32(b, uint32(v)) }

func putF64(b []byte, v float64) {
	binary.LittleEndian.PutUint64(b, math.Float64bits(v))
}
