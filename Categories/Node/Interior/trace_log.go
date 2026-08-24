package interior

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func TraceRelPath(row int) string {
	return fmt.Sprintf("view/nodes/%d/interior-trace.bin", row)
}

func tracePath(sceneRoot string, row int32) string {
	return filepath.Join(sceneRoot, filepath.FromSlash(TraceRelPath(int(row))))
}

var traceEnabled = os.Getenv("BEADNETWORK_PROBE_TRACE") == "1"

func appendTrace(path string, events []RowEvent) {
	if path == "" || len(events) == 0 {
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
