package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: readtrace <trace.bin> [more.bin ...]")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Decodes the binary trace records Go appends beside each item")
		fmt.Fprintln(os.Stderr, "the interior's interior-trace.bin into logfmt lines on stdout.")
		os.Exit(2)
	}

	out := bufio.NewWriter(os.Stdout)

	status := 0
	for _, path := range os.Args[1:] {
		if err := emit(out, path); err != nil {
			fmt.Fprintf(os.Stderr, "readtrace %s: %v\n", path, err)
			status = 1
		}
	}

	if err := out.Flush(); err != nil {
		fmt.Fprintf(os.Stderr, "readtrace: flush: %v\n", err)
		status = 1
	}
	os.Exit(status)
}

func emit(out *bufio.Writer, path string) error {
	buf, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	for off := 0; off < len(buf); {
		e, tsMs, n, ok := DecodeRecord(buf[off:])
		if !ok {
			return fmt.Errorf("truncated record at byte %d of %d", off, len(buf))
		}
		fmt.Fprintln(out, LineOf(e, tsMs))
		off += n
	}
	return nil
}
