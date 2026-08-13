package inputcodec

import "strings"

const InputLayoutFingerprint = "kinds=save:4,raw-input:10,edit-update:22 eventKinds=pointerdown,pointermove,pointerup,wheel,home hitKinds=port,handhold,node,edge,torus,empty updateKinds=overlays,clock,distanceGroup,scene,tiltVector updateAttrs=toggle,speed,length,selected,theta,phi,reset,start,latticePoints,create,delete overlayFlags=tori,scenePoles,nodePoles,selSpherePoles,handholds,labelsGlobal,overlays,nodeBody,nodeRing,ringPick,selectionRing,hoverRing,reachSphere,sceneVectors,commEdges"

const (
	InKindSave = 4

	InKindRawInput = 10

	InKindEditUpdate = 22
)

const (
	InOverlayAttrToggle       = 0
	InClockAttrSpeed          = 1
	InDistanceGroupAttrLength = 2
	InSceneAttrSelected       = 3
	InTiltVectorAttrTheta     = 4

	InTiltVectorAttrReset    = 6
	InTiltVectorAttrStart    = 7
	InSceneAttrLatticePoints = 8

	InSceneAttrCreate = 9
	InSceneAttrDelete = 10
)

var (
	InEventKinds   = parseFPList(InputLayoutFingerprint, "eventKinds=")
	InHitKinds     = parseFPList(InputLayoutFingerprint, "hitKinds=")
	InUpdateKinds  = parseFPList(InputLayoutFingerprint, "updateKinds=")
	InUpdateAttrs  = parseFPList(InputLayoutFingerprint, "updateAttrs=")
	InOverlayFlags = parseFPList(InputLayoutFingerprint, "overlayFlags=")
)

func init() {
	for _, e := range []struct {
		marker string
		list   []string
	}{
		{"eventKinds=", InEventKinds},
		{"hitKinds=", InHitKinds},
		{"updateKinds=", InUpdateKinds},
		{"updateAttrs=", InUpdateAttrs},
		{"overlayFlags=", InOverlayFlags},
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
