package Overlay

type Emitter interface {
	OverlayBreadcrumb(label, scope string, on bool)
	Redraw()
}

type ChannelVectors interface {
	BroadcastChannelVectorsOn(on bool)
}

func EditOverlays(e Edit, ov *OverlayState, chans ChannelVectors, emit Emitter, persist func(OverlayState)) {
	ToggleFlag(ov, chans, emit, e.Flag)
	emit.Redraw()
	persist(*ov)
}

func ToggleFlag(ov *OverlayState, chans ChannelVectors, emit Emitter, flag string) {
	fn, ok := OverlayToggles[flag]
	if !ok {
		return
	}
	fn(ov)
	if flag == "ruleChannels" {
		chans.BroadcastChannelVectorsOn(ov.RuleChannelsVisible)
	}
	if scope, ok := OverlayFlagBreadcrumbScope[flag]; ok {
		emit.OverlayBreadcrumb(BreadcrumbPoleToggleGo, scope, OverlayFlagValue[flag](ov))
	}
}

func SetCount(ov *OverlayState, chans ChannelVectors, emit Emitter, flags []string, target bool) {
	for _, flag := range flags {
		read, ok := OverlayFlagRead[flag]
		if !ok || read(ov) == target {
			continue
		}
		ToggleFlag(ov, chans, emit, flag)
	}
}
