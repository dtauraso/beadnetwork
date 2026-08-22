package trace

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const TraceFile = "trace.bin"

type Owner string

const (
	OwnerView     Owner = "view"
	OwnerNode     Owner = "node"
	OwnerEdge     Owner = "edge"
	OwnerInterior Owner = "interior"
	OwnerBead     Owner = "bead"
)

var sceneRoot string

func SetSceneRoot(dir string) { sceneRoot = dir }

func sceneRootDir() string { return sceneRoot }

func TraceRelPath(owner Owner, row int32) string {
	switch owner {
	case OwnerView:
		return "view/" + TraceFile
	case OwnerEdge:
		return fmt.Sprintf("view/edges/%d/%s", row, TraceFile)
	case OwnerInterior:
		return fmt.Sprintf("view/nodes/%d/interior-%s", row, TraceFile)
	case OwnerBead:
		return fmt.Sprintf("view/nodes/%d/beads-%s", row, TraceFile)
	default:
		return fmt.Sprintf("view/nodes/%d/%s", row, TraceFile)
	}
}

func ownerPath(owner Owner, row int32) string {
	return filepath.Join(sceneRootDir(), filepath.FromSlash(TraceRelPath(owner, row)))
}

var traceEnabled = os.Getenv("WIREFOLD_PROBE_TRACE") == "1"

func TraceEnabled() bool { return traceEnabled }

func LabelOf(id uint8) string {
	if int(id) < len(BreadcrumbLabels) {
		return BreadcrumbLabels[id]
	}
	return fmt.Sprintf("%d", id)
}

var nameOf func(row int32) string

func SetNameResolver(fn func(row int32) string) { nameOf = fn }

func NameOf(row int32) string {
	if row < 0 || nameOf == nil {
		return ""
	}
	return nameOf(row)
}

type Log struct {
	path string
}

func NewLog(owner Owner, row int32) *Log {
	return &Log{path: ownerPath(owner, row)}
}

func (l *Log) open() (*os.File, error) {
	f, err := os.OpenFile(l.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err == nil {
		return f, nil
	}
	if err := os.MkdirAll(filepath.Dir(l.path), 0o755); err != nil {
		return nil, err
	}
	return os.OpenFile(l.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
}

func (l *Log) Append(events []RowEvent) {
	if l == nil || len(events) == 0 {
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

	f, err := l.open()
	if err != nil {
		return
	}
	_, _ = f.Write(out)
	_ = f.Close()
}
