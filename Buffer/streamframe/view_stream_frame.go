package streamframe

import (
	"encoding/binary"

	B "github.com/dtauraso/wirefold/Buffer"
)

func BuildViewStreamFrame(tick uint32,
	camPX, camPY, camPZ, camR, camPosTheta, camPosPhi, camUpTheta, camUpPhi float32,
	overlay B.OverlayRow,
	sceneCX, sceneCY, sceneCZ, sceneRadius float32,
	tabNames []string, tabSelected uint16,
	events []StreamEvent,
) []byte {
	buf := make([]byte, B.BufViewFrameHeaderSize+B.BufCameraStride+B.BufOverlayStride+B.BufSceneStride)
	binary.LittleEndian.PutUint32(buf[0:], tick)
	off := B.BufViewFrameHeaderSize
	B.SetCameraRow(buf[off:], camPX, camPY, camPZ, camR, camPosTheta, camPosPhi, camUpTheta, camUpPhi)
	off += B.BufCameraStride
	B.SetOverlayRow(buf[off:], overlay)
	off += B.BufOverlayStride
	B.SetSceneRow(buf[off:], sceneCX, sceneCY, sceneCZ, sceneRadius)
	buf = append(buf, BuildSceneTabsSection(tabNames, tabSelected)...)
	return append(buf, BuildEventsSection(events)...)
}

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
