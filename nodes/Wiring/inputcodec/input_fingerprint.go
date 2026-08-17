package inputcodec

import "strings"

const InputLayoutFingerprint = "kinds=save:4,raw-input:10,edit-update:22 eventKinds=pointerdown,pointermove,pointerup,wheel,home hitKinds=port,handhold,node,edge,torus,empty updateKinds=overlays,clock,distanceGroup,scene,tiltVector,panels,node,edge updateAttrs=toggle,speed,length,selected,phi,reset,start,latticePoints,create,delete,dragPhi,dragMaxTheta,dragActive,kindActive,selfDragPhi,selfDragMaxTheta,selfDragActive,dragR,selfDragR overlayFlags=tori,scenePoles,nodePoles,handholds,labelsGlobal,overlays,nodeBody,nodeRing,ringPick,selectionRing,hoverRing,sceneVectors,ruleChannels,nodePoleSphere panelFlags=overlays,node,nodeShape,nodeState,nodePoles,nodeRules,scene,sceneGuides,scenePoles,sceneVectors,sceneLabels"

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
	InTiltVectorAttrPhi       = 4

	InTiltVectorAttrReset    = 6
	InTiltVectorAttrStart    = 7
	InSceneAttrLatticePoints = 8

	InSceneAttrCreate = 9
	InSceneAttrDelete = 10

	InPanelAttrToggle = 11

	InNodeAttrDragPhi      = 12
	InNodeAttrDragMaxTheta = 13
	InNodeAttrDragActive   = 14
	InNodeAttrKindActive   = 15

	InNodeAttrSelfDragPhi      = 16
	InNodeAttrSelfDragMaxTheta = 17
	InNodeAttrSelfDragActive   = 18

	InNodeAttrDragR     = 19
	InNodeAttrSelfDragR = 20
)

var (
	InEventKinds   = parseFPList(InputLayoutFingerprint, "eventKinds=")
	InHitKinds     = parseFPList(InputLayoutFingerprint, "hitKinds=")
	InUpdateKinds  = parseFPList(InputLayoutFingerprint, "updateKinds=")
	InUpdateAttrs  = parseFPList(InputLayoutFingerprint, "updateAttrs=")
	InOverlayFlags = parseFPList(InputLayoutFingerprint, "overlayFlags=")
	InPanelFlags   = parseFPList(InputLayoutFingerprint, "panelFlags=")
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
		{"panelFlags=", InPanelFlags},
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
