package interior

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
)

type InteriorStream struct {
	buildFrame func(tick uint32)
	tracePath  string
	tick       uint32

	nodeRow int32

	lastPresent            []uint8
	lastValue              []int32
	lastOx, lastOy, lastOz []float32

	values *ValueWriter
}

func NewInteriorStream(buildFrame func(tick uint32), nodeRow int32, slots int) *InteriorStream {
	absent := make([]uint8, slots)
	zeroI := make([]int32, slots)
	zeroF := make([]float32, slots)
	return &InteriorStream{
		buildFrame: buildFrame, nodeRow: nodeRow,
		lastPresent: absent, lastValue: zeroI,
		lastOx: zeroF, lastOy: append([]float32{}, zeroF...), lastOz: append([]float32{}, zeroF...),
	}
}

func (s *InteriorStream) NodeRowOf() int32 { return s.nodeRow }

func (s *InteriorStream) write(present []uint8, value []int32, ox, oy, oz []float32, events []RowEvent, center Vec3) {
	if s == nil {
		return
	}
	s.lastPresent, s.lastValue = present, value
	s.lastOx, s.lastOy, s.lastOz = ox, oy, oz
	s.tick++
	wx, wy, wz := worldOf(ox, oy, oz, center)

	if s.values != nil {
		s.values.Begin()
		s.values.Bytes("present", present)
		s.values.Bytes("value", packI32(value))
		s.values.Bytes("x", packF32(wx))
		s.values.Bytes("y", packF32(wy))
		s.values.Bytes("z", packF32(wz))
		if err := s.values.Flush(); err != nil {
			fmt.Fprintf(os.Stderr, "interior values write (node row %d): %v\n", s.nodeRow, err)
		}
	}
	appendTrace(s.tracePath, events)
	if s.buildFrame != nil {
		s.buildFrame(s.tick)
	}
}

func packI32(v []int32) []byte {
	b := make([]byte, 0, len(v)*4)
	for _, x := range v {
		b = binary.LittleEndian.AppendUint32(b, uint32(x))
	}
	return b
}

func packF32(v []float32) []byte {
	b := make([]byte, 0, len(v)*4)
	for _, x := range v {
		b = binary.LittleEndian.AppendUint32(b, math.Float32bits(x))
	}
	return b
}

func (s *InteriorStream) SetValueWriter(w *ValueWriter) { s.values = w }

func (s *InteriorStream) SetSceneRoot(sceneRoot string) {
	s.tracePath = tracePath(sceneRoot, s.nodeRow)
}

func worldOf(ox, oy, oz []float32, center Vec3) ([]float32, []float32, []float32) {
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

func (s *InteriorStream) WriteFull(present []uint8, value []int32, ox, oy, oz []float32, events []RowEvent, center Vec3) {
	s.write(present, value, ox, oy, oz, events, center)
}

func (s *InteriorStream) WriteEvents(events []RowEvent, center Vec3) {
	if s == nil {
		return
	}
	s.write(s.lastPresent, s.lastValue, s.lastOx, s.lastOy, s.lastOz, events, center)
}

func (s *InteriorStream) RewriteAtCenter(center Vec3) {
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
