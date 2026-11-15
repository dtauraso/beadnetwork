package Flags

type Emitter interface {
	OverlayBreadcrumb(label, scope string, on bool)
	Redraw()
}

func EditOverlays(e Edit, ov *OverlayState, emit Emitter, persist func(OverlayState)) {
	ToggleFlag(ov, emit, e.Flag)
	emit.Redraw()
	persist(*ov)
}

func ToggleFlag(ov *OverlayState, emit Emitter, flag string) {
	fn, ok := OverlayToggles[flag]
	if !ok {
		return
	}
	fn(ov)
	if scope, ok := OverlayFlagBreadcrumbScope[flag]; ok {
		emit.OverlayBreadcrumb(BreadcrumbPoleToggleGo, scope, OverlayFlagValue[flag](ov))
	}
}

func SetCount(ov *OverlayState, emit Emitter, flags []string, target bool) {
	for _, flag := range flags {
		read, ok := OverlayFlagRead[flag]
		if !ok || read(ov) == target {
			continue
		}
		ToggleFlag(ov, emit, flag)
	}
}

func ApplyUpdate(attr byte, payload []byte, ov *OverlayState, emit Emitter, persist func(OverlayState)) {
	e, ok := DecodeUpdate(payload, attr)
	if !ok {
		return
	}
	EditOverlays(e, ov, emit, persist)
}
