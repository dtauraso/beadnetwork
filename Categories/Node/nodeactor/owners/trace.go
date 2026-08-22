package owners

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Trace struct {
	path   string
	buffer chan []RowEvent
}

const traceBufferDepth = 64

func TraceRelPath(row int) string {
	return fmt.Sprintf("view/nodes/%d/trace.bin", row)
}

func (t *Trace) Wire(sceneRoot string, row int32) {
	t.path = filepath.Join(sceneRoot, filepath.FromSlash(TraceRelPath(int(row))))
	if t.buffer == nil {
		t.buffer = make(chan []RowEvent, traceBufferDepth)
	}
}

func (t *Trace) Post(events []RowEvent) {
	if t.buffer == nil {
		return
	}
	select {
	case t.buffer <- events:
	default:
	}
}

func (t *Trace) Drain() []RowEvent {
	var out []RowEvent
	for {
		select {
		case ev := <-t.buffer:
			out = append(out, ev...)
		default:
			return out
		}
	}
}

var traceEnabled = os.Getenv("WIREFOLD_PROBE_TRACE") == "1"

func (t *Trace) Append(events []RowEvent) {
	if t.path == "" || len(events) == 0 {
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
	f, err := openAppend(t.path)
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
