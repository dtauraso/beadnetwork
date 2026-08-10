package buflayout

import (
	"fmt"
	"strings"
)

// goWriterCall returns the Go snippet to write one column into buf[off+c.offset:].
func goWriterCall(c bufCol) string {
	off := fmt.Sprintf("off+%d", c.offset)
	param := strings.ToLower(c.name[:1]) + c.name[1:]
	return "\t" + lookupBufType(c.bufType).goWrite(off, param)
}

// goParamType returns the Go parameter type for a buf: type tag.
func goParamType(t string) string {
	return lookupBufType(t).goType
}

// tsDataViewGetter returns the DataView getter method for a buf: type.
func tsDataViewGetter(t string) string {
	return lookupBufType(t).tsGetter
}

// tsDataViewLE returns ", true" for multi-byte types (little-endian flag) or ""
// for single-byte types.
func tsDataViewLE(t string) string {
	if lookupBufType(t).tsLE {
		return ", true"
	}
	return ""
}

// colGoName converts a block name + column name to the Go const name.
// e.g. Bead + X → BufBeadColX; Bead + stride → BufBeadStride.
func colGoName(block, col string) string {
	return "Buf" + block + "Col" + col
}

// strideGoName returns the Go stride constant name for a block.
func strideGoName(block string) string {
	return "Buf" + block + "Stride"
}

// colTSName converts a block name + column name to the TS const name (SCREAMING_SNAKE).
// e.g. Bead + X → BEAD_COL_X; Node + CX → NODE_COL_CX.
func colTSName(block, col string) string {
	return camelToScreamingSnake(block) + "_COL_" + camelToScreamingSnake(col)
}

// strideTSName returns the TS stride constant name for a block.
func strideTSName(block string) string {
	return camelToScreamingSnake(block) + "_STRIDE"
}

// writerFnGoName returns the Go writer function name for a block.
func writerFnGoName(block string) string {
	return "Set" + block + "Row"
}

// readerFnTSName returns the TS reader function name for one column.
func readerFnTSName(block, col string) string {
	return "read" + block + col
}
