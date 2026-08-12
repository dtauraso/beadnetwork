package buflayout

import (
	"bufio"
	"bytes"
	"os"
)

func writeBufferLayoutTSRows(outPathA, outPathB string, schema BufLayoutSchema) error {
	blocks := rowBlocks(schema)
	split := rowBlockSplit(len(blocks))
	if err := writeBufferLayoutTSRowsFile(outPathA, blocks[:split]); err != nil {
		return err
	}
	return writeBufferLayoutTSRowsFile(outPathB, blocks[split:])
}

func writeBufferLayoutTSRowsFile(outPath string, blocks []bufBlock) error {
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)

	for _, blk := range blocks {
		writeTSBlockConstsAndReaders(w, blk, false)
	}

	w.Flush()
	return os.WriteFile(outPath, buf.Bytes(), 0644)
}
