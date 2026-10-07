package Drag

import "slices"

const KindRawInput = 10

var EventKinds = []string{
	"pointerdown",
	"pointermove",
	"pointerup",
	"wheel",
	"home",
	"delete",
	"key",
}

var CommandKinds = []string{
	"delete",
	"key",
}

func IsCommandKind(kind string) bool { return slices.Contains(CommandKinds, kind) }

var HitKinds = []string{
	"port",
	"handhold",
	"node",
	"edge",
	"torus",
	"empty",
}
