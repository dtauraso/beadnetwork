package main

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

var traceEventKinds = []string{KindRecv, KindFire, KindSend, KindArrive, KindBreadcrumb}

const traceRecordFixedSize = 82

func kindOf(id uint8) string {
	if id == 0 || int(id) > len(traceEventKinds) {
		return ""
	}
	return traceEventKinds[id-1]
}

func DecodeRecord(buf []byte) (RowEvent, int64, int, bool) {
	if len(buf) < traceRecordFixedSize {
		return RowEvent{}, 0, 0, false
	}
	labelLen := int(buf[2])
	textLen := int(binary.LittleEndian.Uint16(buf[80:]))
	total := traceRecordFixedSize + labelLen + textLen
	if len(buf) < total {
		return RowEvent{}, 0, 0, false
	}
	labelEnd := traceRecordFixedSize + labelLen

	e := RowEvent{
		Kind:          kindOf(buf[0]),
		Debug:         buf[1],
		Label:         string(buf[traceRecordFixedSize:labelEnd]),
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
		Text:          string(buf[labelEnd:total]),
	}
	return e, int64(binary.LittleEndian.Uint64(buf[32:])), total, true
}

func getI32(b []byte) int32 { return int32(binary.LittleEndian.Uint32(b)) }

func getF64(b []byte) float64 {
	return math.Float64frombits(binary.LittleEndian.Uint64(b))
}

func NameOf(row int32) string { return "" }
