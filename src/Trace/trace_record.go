package trace

import (
	"encoding/binary"
	"math"
)

const TraceRecordVersion = 1

const TraceRecordFixedSize = 82

func KindID(kind string) uint8 {
	for i, k := range TraceEventKinds {
		if k == kind {
			return uint8(i + 1)
		}
	}
	return 0
}

func KindOf(id uint8) string {
	if id == 0 || int(id) > len(TraceEventKinds) {
		return ""
	}
	return TraceEventKinds[id-1]
}

func AppendRecord(buf []byte, e RowEvent, nowMs int64) []byte {
	var rec [TraceRecordFixedSize]byte

	rec[0] = KindID(e.Kind)
	rec[1] = e.Label
	rec[2] = e.Debug

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
	return append(buf, text...)
}

func DecodeRecord(buf []byte) (RowEvent, int64, int, bool) {
	if len(buf) < TraceRecordFixedSize {
		return RowEvent{}, 0, 0, false
	}
	textLen := int(binary.LittleEndian.Uint16(buf[80:]))
	total := TraceRecordFixedSize + textLen
	if len(buf) < total {
		return RowEvent{}, 0, 0, false
	}

	e := RowEvent{
		Kind:          KindOf(buf[0]),
		Label:         buf[1],
		Debug:         buf[2],
		NodeRow:       getI32(buf[4:]),
		PortRow:       getI32(buf[8:]),
		TargetRow:     getI32(buf[12:]),
		TargetPortRow: getI32(buf[16:]),
		EdgeRow:       getI32(buf[20:]),
		Slot:          getI32(buf[24:]),
		Value:         getI32(buf[28:]),
		Bead:          binary.LittleEndian.Uint64(buf[40:]),
		BeadSteps:     getF64(buf[48:]),
		X:             getF64(buf[56:]),
		Y:             getF64(buf[64:]),
		Z:             getF64(buf[72:]),
		Text:          string(buf[TraceRecordFixedSize:total]),
	}
	return e, int64(binary.LittleEndian.Uint64(buf[32:])), total, true
}

func putI32(b []byte, v int32) { binary.LittleEndian.PutUint32(b, uint32(v)) }

func getI32(b []byte) int32 { return int32(binary.LittleEndian.Uint32(b)) }

func putF64(b []byte, v float64) {
	binary.LittleEndian.PutUint64(b, math.Float64bits(v))
}

func getF64(b []byte) float64 {
	return math.Float64frombits(binary.LittleEndian.Uint64(b))
}
