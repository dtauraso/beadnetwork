package buflayout

import (
	"fmt"
	"strings"
)

func goWriterCall(c bufCol) string {
	off := fmt.Sprintf("off+%d", c.offset)
	param := strings.ToLower(c.name[:1]) + c.name[1:]
	return "\t" + lookupBufType(c.bufType).goWrite(off, param)
}

func goParamType(t string) string {
	return lookupBufType(t).goType
}

func tsDataViewGetter(t string) string {
	return lookupBufType(t).tsGetter
}

func tsDataViewLE(t string) string {
	if lookupBufType(t).tsLE {
		return ", true"
	}
	return ""
}

func colGoName(block, col string) string {
	return "Buf" + block + "Col" + col
}

func strideGoName(block string) string {
	return "Buf" + block + "Stride"
}

func colTSName(block, col string) string {
	return camelToScreamingSnake(block) + "_COL_" + camelToScreamingSnake(col)
}

func strideTSName(block string) string {
	return camelToScreamingSnake(block) + "_STRIDE"
}

func writerFnGoName(block string) string {
	return "Set" + block + "Row"
}

func readerFnTSName(block, col string) string {
	return "read" + block + col
}
