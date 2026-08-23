package View

import (
	"os"
	"path/filepath"
	"time"
)

func TraceRelPath() string { return "view/trace.bin" }

func viewTracePath(sceneRoot string) string {
	return filepath.Join(sceneRoot, filepath.FromSlash(TraceRelPath()))
}

func (ui *UIState) SceneRoot() string { return ui.sceneRoot }

func (ui *UIState) Redraw() { ui.EmitViewFrame(nil) }

func (ui *UIState) OverlayBreadcrumb(label, scope string, on bool) {
	var v int32
	if on {
		v = 1
	}
	ui.EmitBreadcrumb(RowEvent{
		Label: label, NodeRow: -1, PortRow: -1, TargetRow: -1,
		TargetPortRow: -1, EdgeRow: -1, Slot: -1,
		Value: v, Text: scope,
	})
}

var traceEnabled = os.Getenv("WIREFOLD_PROBE_TRACE") == "1"

func appendTrace(path string, events []RowEvent) {
	if len(events) == 0 {
		return
	}
	var out []byte
	now := time.Now().UnixMilli()
	for _, e := range events {
		if !traceEnabled && e.Kind != KindBreadcrumb {
			continue
		}
		out = AppendRecord(out, e, now)
	}
	if len(out) == 0 {
		return
	}
	f, err := openAppend(path)
	if err != nil {
		return
	}
	_, _ = f.Write(out)
	_ = f.Close()
}

func openAppend(path string) (*os.File, error) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err == nil {
		return f, nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	return os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
}
