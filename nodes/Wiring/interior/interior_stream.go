package interior

import (
	"encoding/binary"
	"io"

	"github.com/dtauraso/wirefold/nodes/rowevent"
)

type InteriorStream struct {
	out        io.Writer
	buildFrame func(tick uint32, present []uint8, value []int32, ox, oy, oz []float32, events []rowevent.RowEvent) []byte
	tick       uint32

	nodeRow int32

	lastPresent            []uint8
	lastValue              []int32
	lastOx, lastOy, lastOz []float32
}

func NewInteriorStream(out io.Writer, buildFrame func(tick uint32, present []uint8, value []int32, ox, oy, oz []float32, events []rowevent.RowEvent) []byte, nodeRow int32, slots int) *InteriorStream {
	absent := make([]uint8, slots)
	zeroI := make([]int32, slots)
	zeroF := make([]float32, slots)
	return &InteriorStream{
		out: out, buildFrame: buildFrame, nodeRow: nodeRow,
		lastPresent: absent, lastValue: zeroI,
		lastOx: zeroF, lastOy: append([]float32{}, zeroF...), lastOz: append([]float32{}, zeroF...),
	}
}

func (s *InteriorStream) OutWriter() io.Writer { return s.out }

func (s *InteriorStream) NodeRowOf() int32 { return s.nodeRow }

func (s *InteriorStream) write(present []uint8, value []int32, ox, oy, oz []float32, events []rowevent.RowEvent, center vec3) {
	if s == nil {
		return
	}
	s.lastPresent, s.lastValue = present, value
	s.lastOx, s.lastOy, s.lastOz = ox, oy, oz
	s.tick++
	wx, wy, wz := worldOf(ox, oy, oz, center)
	writeInteriorStreamFrame(s.out, s.buildFrame, s.tick, present, value, wx, wy, wz, events)
}

func worldOf(ox, oy, oz []float32, center vec3) ([]float32, []float32, []float32) {
	wx := make([]float32, len(ox))
	wy := make([]float32, len(oy))
	wz := make([]float32, len(oz))
	for i := range ox {
		wx[i] = ox[i] + float32(center.X)
	}
	for i := range oy {
		wy[i] = oy[i] + float32(center.Y)
	}
	for i := range oz {
		wz[i] = oz[i] + float32(center.Z)
	}
	return wx, wy, wz
}

func (s *InteriorStream) WriteFull(present []uint8, value []int32, ox, oy, oz []float32, events []rowevent.RowEvent, center vec3) {
	s.write(present, value, ox, oy, oz, events, center)
}

func (s *InteriorStream) WriteEvents(events []rowevent.RowEvent, center vec3) {
	if s == nil {
		return
	}
	s.write(s.lastPresent, s.lastValue, s.lastOx, s.lastOy, s.lastOz, events, center)
}

func (s *InteriorStream) RewriteAtCenter(center vec3) {
	if s == nil {
		return
	}
	s.write(s.lastPresent, s.lastValue, s.lastOx, s.lastOy, s.lastOz, nil, center)
}

func boolU8(b bool) uint8 {
	if b {
		return 1
	}
	return 0
}

func writeInteriorStreamFrame(out io.Writer, buildFrame func(tick uint32, present []uint8, value []int32, ox, oy, oz []float32, events []rowevent.RowEvent) []byte, tick uint32, present []uint8, value []int32, ox, oy, oz []float32, events []rowevent.RowEvent) {
	if out == nil || buildFrame == nil {
		return
	}
	frame := buildFrame(tick, present, value, ox, oy, oz, events)

	buf := make([]byte, 4+len(frame))
	binary.LittleEndian.PutUint32(buf[:4], uint32(len(frame)))
	copy(buf[4:], frame)

	_, _ = out.Write(buf)
}
