package Overlay

import (
	"github.com/dtauraso/wirefold/src/Input/Stdin"
	T "github.com/dtauraso/wirefold/src/Trace"
)

type Emitter interface {
	EmitBreadcrumb(ev T.RowEvent)
	EmitViewFrame(events []T.RowEvent)
}

type ChannelVectors interface {
	BroadcastChannelVectorsOn(on bool)
}

func EditOverlays(msg Stdin.StdinMsg, ov *OverlayState, chans ChannelVectors, emit Emitter, persist func(OverlayState)) {
	if msg.Attr != "toggle" {
		return
	}
	ToggleFlag(ov, chans, emit, msg.Flag)
	emit.EmitViewFrame(nil)
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
		var v uint8
		if OverlayFlagValue[flag](ov) {
			v = 1
		}
		emit.EmitBreadcrumb(T.RowEvent{
			Label: T.BreadcrumbPoleToggleGo, NodeRow: -1, PortRow: -1, TargetRow: -1,
			TargetPortRow: -1, EdgeRow: -1, Slot: -1,
			Value: int32(v), Text: scope,
		})
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
