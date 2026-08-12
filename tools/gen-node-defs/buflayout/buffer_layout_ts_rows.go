package buflayout

import (
	"bufio"
	"bytes"
	"os"
)

func writeBufferLayoutTSRows(outPath string, schema BufLayoutSchema) error {
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)

	for _, blk := range schema.Blocks {
		if isSingletonBlock(blk.name) {
			continue
		}
		writeTSBlockConstsAndReaders(w, blk, false)
	}

	w.Flush()
	return os.WriteFile(outPath, buf.Bytes(), 0644)
}
