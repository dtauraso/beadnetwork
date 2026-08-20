package stdinreader

import (
	"context"
	"fmt"
	"github.com/dtauraso/wirefold/src/Chrome/SliderPanel"

	"github.com/dtauraso/wirefold/src/Node/Wiring/dispatch"
	"github.com/dtauraso/wirefold/src/Node/Wiring/inputcodec"
	T "github.com/dtauraso/wirefold/src/Trace"
)

func applyUpdateEdge(ctx context.Context, msg inputcodec.StdinMsg, md *dispatch.MoveDispatch, tr *T.Trace, speedSinks SliderPanel.Sinks) {
	if md == nil || msg.Attr != "dragActive" {
		return
	}
	row := msg.Num
	if row < 0 || row >= len(md.Rules.TogglesByEdgeRow) {
		panic(fmt.Sprintf(
			"applyUpdateEdge: edge row %d is outside the %d rows the tree declares, so a rule toggle names an edge "+
				"the row space has no slot for — the webview and the loaded tree disagree about how many edges exist",
			row, len(md.Rules.TogglesByEdgeRow)))
	}
	toggle := md.Rules.TogglesByEdgeRow[row]
	if toggle == nil {
		return
	}
	select {
	case toggle <- struct{}{}:
	case <-ctx.Done():
	}
}
