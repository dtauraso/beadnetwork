package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: readgrid <grid.bin> [more.bin ...]")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Prints a node's grid.bin as the table the Node grid tab shows: one column per step.")
		os.Exit(2)
	}
	status := 0
	for _, path := range os.Args[1:] {
		if err := dump(path); err != nil {
			fmt.Fprintf(os.Stderr, "readgrid %s: %v\n", path, err)
			status = 1
		}
	}
	os.Exit(status)
}

func num(v float64) string {
	return strconv.FormatFloat(float64(int64(v*1000+copysign(v)))/1000, 'f', -1, 64)
}

func copysign(v float64) float64 {
	if v < 0 {
		return -0.5
	}
	return 0.5
}

func dump(path string) error {
	h, rows, err := Links.ReadGridRows(path)
	if err != nil {
		return err
	}
	fmt.Printf("%s  node %s  level %q  shown step %d\n", path, h.ID, h.Level, h.Shown)
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	head := []string{""}
	for _, s := range h.Steps {
		head = append(head, "step "+strconv.Itoa(s))
	}
	fmt.Fprintln(w, strings.Join(head, "\t"))
	band := ""
	for _, r := range rows {
		if r.Band != band {
			band = r.Band
			fmt.Fprintln(w, "["+band+"]")
		}
		cells := []string{r.Name}
		for _, v := range r.V {
			cells = append(cells, num(v))
		}
		fmt.Fprintln(w, strings.Join(cells, "\t"))
	}
	band = ""
	for _, r := range h.Base {
		if r.Band != band {
			band = r.Band
			fmt.Fprintln(w, "[sequence "+band+"]")
		}
		fmt.Fprintln(w, r.Name+"\t"+num(r.V[0]))
	}
	return w.Flush()
}
