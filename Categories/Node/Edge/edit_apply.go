package edge

import (
	"context"
	"fmt"
)

func EditEdge(ctx context.Context, e Edit, toggles []chan<- struct{}) {
	ToggleDragActive(ctx, e.Num, toggles)
}

func ToggleDragActive(ctx context.Context, row int, toggles []chan<- struct{}) {
	if row < 0 || row >= len(toggles) {
		panic(fmt.Sprintf(
			"edge.ToggleDragActive: edge row %d is outside the %d rows the tree declares, so a rule toggle names an edge "+
				"the row space has no slot for — the webview and the loaded tree disagree about how many edges exist",
			row, len(toggles)))
	}
	toggle := toggles[row]
	if toggle == nil {
		return
	}
	select {
	case toggle <- struct{}{}:
	case <-ctx.Done():
	}
}
