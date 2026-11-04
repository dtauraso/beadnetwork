package Sheet

import (
	"strings"
	"unicode/utf8"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Panel"
)

func wrapped(g Group, room float32) Group {
	if !g.Wraps {
		return g
	}
	fit := int((room - 2*CellPadX) / Panel.Advance(KeyFontPx))
	var rows [][]Cell
	for _, r := range g.Rows {
		if len(r) != 1 || r[0].Kind != CellText || fit < 1 {
			rows = append(rows, r)
			continue
		}
		line := ""
		for w := range strings.FieldsSeq(r[0].Key) {
			if line != "" && utf8.RuneCountInString(line)+1+utf8.RuneCountInString(w) > fit {
				rows = append(rows, []Cell{Text(line)})
				line = ""
			}
			if line != "" {
				line += " "
			}
			line += w
		}
		rows = append(rows, []Cell{Text(line)})
	}
	g.Rows = rows
	return g
}
