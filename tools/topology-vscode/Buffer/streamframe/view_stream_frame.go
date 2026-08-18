package streamframe

import (
	"encoding/binary"
	"fmt"

	B "github.com/dtauraso/wirefold/tools/topology-vscode/Buffer"
)

func BuildViewStreamFrame(tick uint32,
	camPX, camPY, camPZ, camR, camPosPhi, camPosTheta, camUpPhi, camUpTheta float32,
	overlay B.OverlayRow,
	panel B.PanelRow,
	sceneCX, sceneCY, sceneCZ, sceneRadius float32,
	ringSurfacePoints []float32,
	beadRingSurfacePoints []float32,
	tabNames []string, tabSelected uint16,
	events []StreamEvent,
) []byte {
	if len(ringSurfacePoints)%3 != 0 {
		panic(fmt.Sprintf(
			"BuildViewStreamFrame: ringSurfacePoints has %d floats — must be a whole number of XYZ triples",
			len(ringSurfacePoints)))
	}
	if len(beadRingSurfacePoints)%3 != 0 {
		panic(fmt.Sprintf(
			"BuildViewStreamFrame: beadRingSurfacePoints has %d floats — must be a whole number of XYZ triples",
			len(beadRingSurfacePoints)))
	}
	ringPointCount := len(ringSurfacePoints) / 3
	ringSurfaceSize := ringPointCount * B.BufRingPointStride
	beadRingPointCount := len(beadRingSurfacePoints) / 3
	beadRingSurfaceSize := beadRingPointCount * B.BufRingPointStride

	buf := make([]byte, B.BufViewFrameHeaderSize+B.BufCameraStride+B.BufOverlayStride+B.BufPanelStride+B.BufSceneStride+ringSurfaceSize+beadRingSurfaceSize)
	binary.LittleEndian.PutUint32(buf[0:], tick)
	binary.LittleEndian.PutUint32(buf[4:], B.BufLayoutFingerprintHash)
	off := B.BufViewFrameHeaderSize
	B.SetCameraRow(buf[off:], camPX, camPY, camPZ, camR, camPosPhi, camPosTheta, camUpPhi, camUpTheta)
	off += B.BufCameraStride
	B.SetOverlayRow(buf[off:], overlay)
	off += B.BufOverlayStride
	B.SetPanelRow(buf[off:], panel)
	off += B.BufPanelStride
	B.SetSceneRow(buf[off:], sceneCX, sceneCY, sceneCZ, sceneRadius)
	off += B.BufSceneStride

	for i := 0; i < ringPointCount; i++ {
		rowOff := off + i*B.BufRingPointStride
		B.SetRingPointRow(buf[rowOff:rowOff+B.BufRingPointStride], 0, ringSurfacePoints[i*3], ringSurfacePoints[i*3+1], ringSurfacePoints[i*3+2])
	}
	off += ringSurfaceSize

	for i := 0; i < beadRingPointCount; i++ {
		rowOff := off + i*B.BufRingPointStride
		B.SetRingPointRow(buf[rowOff:rowOff+B.BufRingPointStride], 0, beadRingSurfacePoints[i*3], beadRingSurfacePoints[i*3+1], beadRingSurfacePoints[i*3+2])
	}
	off += beadRingSurfaceSize

	if off != len(buf) {
		panic(fmt.Sprintf(
			"BuildViewStreamFrame: packed %d bytes but allocated %d — the fixed-section walk and the size formula disagree",
			off, len(buf)))
	}

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
