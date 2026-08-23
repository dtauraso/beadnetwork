package main

import "strings"

const InputLayoutFingerprint = "kinds=save:4,raw-input:10,edit-update:22 eventKinds=pointerdown,pointermove,pointerup,wheel,home,delete,key hitKinds=port,handhold,node,edge,torus,empty updateKinds=overlays,clock,scene,tiltVector,panels,node,edge"

const (
	InKindSave = 4

	InKindRawInput = 10

	InKindEditUpdate = 22
)

var (
	InEventKinds  = parseFPList(InputLayoutFingerprint, "eventKinds=")
	InHitKinds    = parseFPList(InputLayoutFingerprint, "hitKinds=")
	InUpdateKinds = parseFPList(InputLayoutFingerprint, "updateKinds=")
)

func init() {
	for _, e := range []struct {
		marker string
		list   []string
	}{
		{"eventKinds=", InEventKinds},
		{"hitKinds=", InHitKinds},
		{"updateKinds=", InUpdateKinds},
	} {
		if len(e.list) == 0 {
			panic("input_codec: INPUT_LAYOUT_FINGERPRINT is missing the " + e.marker + " token — the wire enum orderings derive from it")
		}
	}
}

func parseFPList(fp, marker string) []string {
	i := strings.Index(fp, marker)
	if i < 0 {
		return nil
	}
	rest := fp[i+len(marker):]
	if sp := strings.IndexByte(rest, ' '); sp >= 0 {
		rest = rest[:sp]
	}
	return strings.Split(rest, ",")
}
