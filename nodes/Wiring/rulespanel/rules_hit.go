package rulespanel

import (
	"strconv"
	"strings"

	"github.com/dtauraso/wirefold/nodes/Wiring/panelstack"
)

type HitKind int

const (
	HitNone HitKind = iota
	HitToggle
	HitCheck
	HitValue
	HitShared
	HitMenuRow
)

type Hit struct {
	Kind HitKind

	Check CheckKind
	Value ValueKind

	NodeRow int32
	EdgeRow int32
}

func (l Layout) Hit(x, y float64) Hit {
	if panelstack.HitRect(l.Toggle, x, y) {
		return Hit{Kind: HitToggle}
	}
	if !l.Open {
		return Hit{}
	}
	if l.MenuOpen {
		for _, m := range l.MenuRows {
			if panelstack.HitRect(m.Rect, x, y) {
				return Hit{Kind: HitMenuRow, NodeRow: m.NodeRow}
			}
		}
	}
	for _, r := range l.Rows {
		if r.Check != CheckNone && panelstack.HitRect(r.CheckRect, x, y) {
			return Hit{Kind: HitCheck, Check: r.Check, NodeRow: r.NodeRow, EdgeRow: r.EdgeRow}
		}
		if r.Kind == RowNodeHead && panelstack.HitRect(r.SharedRect, x, y) {
			return Hit{Kind: HitShared, NodeRow: r.NodeRow}
		}
		if r.Kind == RowLine && r.Value != ValNone && panelstack.HitRect(r.ValueRect, x, y) {
			return Hit{Kind: HitValue, Value: r.Value, NodeRow: r.NodeRow, EdgeRow: r.EdgeRow}
		}
	}
	return Hit{}
}

func IsTheta(v ValueKind) bool { return v == ValSelfTheta || v == ValDragTheta }

func ParsePiDraft(s string) (float64, bool) {
	s = strings.TrimSpace(s)
	slash := strings.Index(s, "/")
	if slash < 0 {
		v, err := strconv.ParseFloat(s, 64)
		return v, err == nil
	}
	p, errP := strconv.ParseFloat(strings.TrimSpace(s[:slash]), 64)
	q, errQ := strconv.ParseFloat(strings.TrimSpace(s[slash+1:]), 64)
	if errP != nil || errQ != nil || q == 0 {
		return 0, false
	}
	return p / q, true
}
