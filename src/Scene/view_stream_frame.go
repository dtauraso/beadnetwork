package Scene

import (
	"encoding/binary"

	B "github.com/dtauraso/wirefold/src/Buffer"
)

func BuildViewStreamFrame(tick uint32) []byte {
	buf := make([]byte, B.BufViewFrameHeaderSize)
	binary.LittleEndian.PutUint32(buf[0:], tick)
	binary.LittleEndian.PutUint32(buf[4:], B.BufLayoutFingerprintHash)
	return buf
}
