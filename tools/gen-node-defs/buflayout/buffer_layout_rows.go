package buflayout

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"
)

func writeBufferLayoutGoRows(outPath string, schema BufLayoutSchema, fp string) error {
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)
	writeBufferLayoutGoPreamble(w, fp)

	for _, blk := range schema.Blocks {
		if isSingletonBlock(blk.name) {
			continue
		}
		writeBufferLayoutGoBlockConst(w, blk)

		var params []string
		for _, c := range blk.columns {
			pname := strings.ToLower(c.name[:1]) + c.name[1:]
			params = append(params, fmt.Sprintf("%s %s", pname, goParamType(c.bufType)))
		}
		fmt.Fprintf(w, "// %s writes one %s row into buf[row*%s:].\n", writerFnGoName(blk.name), blk.name, strideGoName(blk.name))
		fmt.Fprintf(w, "func %s(buf []byte, row int, %s) {\n", writerFnGoName(blk.name), strings.Join(params, ", "))
		fmt.Fprintf(w, "\toff := row * %s\n", strideGoName(blk.name))
		for _, c := range blk.columns {
			fmt.Fprintln(w, goWriterCall(c))
		}
		fmt.Fprintln(w, `}`)
	}

	w.Flush()
	return formatAndWrite(buf, outPath)
}

func writeBufferLayoutGoBlockConst(w *bufio.Writer, blk bufBlock) {
	fmt.Fprintln(w)
	fmt.Fprintf(w, "// ── %s block ", blk.name)
	fmt.Fprintln(w, strings.Repeat("─", 60-len(blk.name)-9))
	fmt.Fprintln(w, `const (`)
	for _, c := range blk.columns {
		fmt.Fprintf(w, "\t%-35s = %d // %s\n", colGoName(blk.name, c.name), c.offset, c.bufType)
	}
	fmt.Fprintf(w, "\t%-35s = %d\n", strideGoName(blk.name), blk.stride)
	fmt.Fprintln(w, `)`)
	fmt.Fprintln(w)
}
