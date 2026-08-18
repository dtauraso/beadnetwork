package streamframe

import (
	"encoding/binary"

	"github.com/dtauraso/wirefold/nodes/rowevent"
	B "github.com/dtauraso/wirefold/tools/topology-vscode/Buffer"
	T "github.com/dtauraso/wirefold/tools/topology-vscode/Trace"
)

func BuildEventsSection(events []rowevent.RowEvent) []byte {
	var recv, fire, send, arrive, crumb []rowevent.RowEvent
	for _, e := range events {
		switch e.Kind {
		case T.KindRecv:
			recv = append(recv, e)
		case T.KindFire:
			fire = append(fire, e)
		case T.KindSend:
			send = append(send, e)
		case T.KindArrive:
			arrive = append(arrive, e)
		case T.KindBreadcrumb:
			crumb = append(crumb, e)
		}
	}

	buf := make([]byte, 0, 20)
	buf = appendRecvSection(buf, recv)
	buf = appendFireSection(buf, fire)
	buf = appendSendSection(buf, send)
	buf = appendArriveSection(buf, arrive)
	return appendBreadcrumbSection(buf, crumb)
}

func appendCount(buf []byte, n int) []byte {
	var hdr [4]byte
	binary.LittleEndian.PutUint32(hdr[:], uint32(n))
	return append(buf, hdr[:]...)
}

func appendRecvSection(buf []byte, events []rowevent.RowEvent) []byte {
	buf = appendCount(buf, len(events))
	rows := make([]byte, len(events)*B.BufRecvStride)
	for i, e := range events {
		B.SetRecvRow(rows, i, e.NodeRow, e.Value)
	}
	return append(buf, rows...)
}

func appendFireSection(buf []byte, events []rowevent.RowEvent) []byte {
	buf = appendCount(buf, len(events))
	rows := make([]byte, len(events)*B.BufFireStride)
	for i, e := range events {
		B.SetFireRow(rows, i, e.NodeRow)
	}
	return append(buf, rows...)
}

func appendSendSection(buf []byte, events []rowevent.RowEvent) []byte {
	buf = appendCount(buf, len(events))
	rows := make([]byte, len(events)*B.BufSendStride)
	for i, e := range events {
		B.SetSendRow(rows, i, e.NodeRow, e.TargetRow, e.Value, float32(e.BeadSteps))
	}
	return append(buf, rows...)
}

func appendArriveSection(buf []byte, events []rowevent.RowEvent) []byte {
	buf = appendCount(buf, len(events))
	rows := make([]byte, len(events)*B.BufArriveStride)
	for i, e := range events {
		B.SetArriveRow(rows, i, e.NodeRow, e.Value, uint32(e.Bead))
	}
	return append(buf, rows...)
}

func appendBreadcrumbSection(buf []byte, events []rowevent.RowEvent) []byte {
	textBytes := make([][]byte, len(events))
	textLen := 0
	for i, e := range events {
		textBytes[i] = []byte(e.Text)
		textLen += len(textBytes[i])
	}
	buf = appendCount(buf, len(events))
	rows := make([]byte, len(events)*B.BufBreadcrumbStride)
	textOff := uint32(0)
	for i, e := range events {
		B.SetBreadcrumbRow(rows, i,
			e.NodeRow, e.PortRow, e.TargetRow, e.TargetPortRow, e.EdgeRow, e.Slot, e.Value,
			float32(e.X), float32(e.Y), float32(e.Z),
			e.Label, e.Debug, textOff, uint32(len(textBytes[i])))
		textOff += uint32(len(textBytes[i]))
	}
	buf = append(buf, rows...)
	for _, tb := range textBytes {
		buf = append(buf, tb...)
	}
	return buf
}
