// ── Camera block ─────────────────────────────────────────────
export const CAMERA_COL_PX                       = 0; // f32
export const CAMERA_COL_PY                       = 4; // f32
export const CAMERA_COL_PZ                       = 8; // f32
export const CAMERA_COL_R                        = 12; // f32
export const CAMERA_COL_POS_PHI                  = 16; // f32
export const CAMERA_COL_POS_THETA                = 20; // f32
export const CAMERA_COL_UP_PHI                   = 24; // f32
export const CAMERA_COL_UP_THETA                 = 28; // f32
export const CAMERA_STRIDE                       = 32;

export function readCameraPX(view: DataView): number { return view.getFloat32(CAMERA_COL_PX, true); }
export function readCameraPY(view: DataView): number { return view.getFloat32(CAMERA_COL_PY, true); }
export function readCameraPZ(view: DataView): number { return view.getFloat32(CAMERA_COL_PZ, true); }
export function readCameraR(view: DataView): number { return view.getFloat32(CAMERA_COL_R, true); }
export function readCameraPosPhi(view: DataView): number { return view.getFloat32(CAMERA_COL_POS_PHI, true); }
export function readCameraPosTheta(view: DataView): number { return view.getFloat32(CAMERA_COL_POS_THETA, true); }
export function readCameraUpPhi(view: DataView): number { return view.getFloat32(CAMERA_COL_UP_PHI, true); }
export function readCameraUpTheta(view: DataView): number { return view.getFloat32(CAMERA_COL_UP_THETA, true); }

// ── Overlay block ────────────────────────────────────────────
export const OVERLAY_COL_SCENE_TORI              = 0; // u8
export const OVERLAY_COL_SCENE_POLES             = 1; // u8
export const OVERLAY_COL_NODE_POLES              = 2; // u8
export const OVERLAY_COL_SEL_SPHERE_POLES        = 3; // u8
export const OVERLAY_COL_HANDHOLDS               = 4; // u8
export const OVERLAY_COL_LABELS_GLOBAL           = 5; // u8
export const OVERLAY_COL_OVERLAYS_VIS            = 6; // u8
export const OVERLAY_COL_NODE_BODY               = 7; // u8
export const OVERLAY_COL_NODE_RING               = 8; // u8
export const OVERLAY_COL_RING_PICK               = 9; // u8
export const OVERLAY_COL_SELECTION_RING          = 10; // u8
export const OVERLAY_COL_HOVER_RING              = 11; // u8
export const OVERLAY_COL_REACH_SPHERE            = 12; // u8
export const OVERLAY_COL_SCENE_VECTORS           = 13; // u8
export const OVERLAY_COL_DRAG_NODE_ROW           = 14; // i32
export const OVERLAY_COL_EDIT_REFUSED            = 18; // u32
export const OVERLAY_COL_SCENE_EDITABLE          = 22; // u8
export const OVERLAY_COL_SCENE_KINDS             = 23; // u32
export const OVERLAY_COL_GROUP_LEN_TIME          = 27; // f32
export const OVERLAY_COL_GROUP_LEN_INPUT         = 31; // f32
export const OVERLAY_COL_GROUP_LEN_GATE          = 35; // f32
export const OVERLAY_COL_SPEED                   = 39; // f32
export const OVERLAY_STRIDE                      = 43;

export function readOverlaySceneTori(view: DataView): number { return view.getUint8(OVERLAY_COL_SCENE_TORI); }
export function readOverlayScenePoles(view: DataView): number { return view.getUint8(OVERLAY_COL_SCENE_POLES); }
export function readOverlayNodePoles(view: DataView): number { return view.getUint8(OVERLAY_COL_NODE_POLES); }
export function readOverlaySelSpherePoles(view: DataView): number { return view.getUint8(OVERLAY_COL_SEL_SPHERE_POLES); }
export function readOverlayHandholds(view: DataView): number { return view.getUint8(OVERLAY_COL_HANDHOLDS); }
export function readOverlayLabelsGlobal(view: DataView): number { return view.getUint8(OVERLAY_COL_LABELS_GLOBAL); }
export function readOverlayOverlaysVis(view: DataView): number { return view.getUint8(OVERLAY_COL_OVERLAYS_VIS); }
export function readOverlayNodeBody(view: DataView): number { return view.getUint8(OVERLAY_COL_NODE_BODY); }
export function readOverlayNodeRing(view: DataView): number { return view.getUint8(OVERLAY_COL_NODE_RING); }
export function readOverlayRingPick(view: DataView): number { return view.getUint8(OVERLAY_COL_RING_PICK); }
export function readOverlaySelectionRing(view: DataView): number { return view.getUint8(OVERLAY_COL_SELECTION_RING); }
export function readOverlayHoverRing(view: DataView): number { return view.getUint8(OVERLAY_COL_HOVER_RING); }
export function readOverlayReachSphere(view: DataView): number { return view.getUint8(OVERLAY_COL_REACH_SPHERE); }
export function readOverlaySceneVectors(view: DataView): number { return view.getUint8(OVERLAY_COL_SCENE_VECTORS); }
export function readOverlayDragNodeRow(view: DataView): number { return view.getInt32(OVERLAY_COL_DRAG_NODE_ROW, true); }
export function readOverlayEditRefused(view: DataView): number { return view.getUint32(OVERLAY_COL_EDIT_REFUSED, true); }
export function readOverlaySceneEditable(view: DataView): number { return view.getUint8(OVERLAY_COL_SCENE_EDITABLE); }
export function readOverlaySceneKinds(view: DataView): number { return view.getUint32(OVERLAY_COL_SCENE_KINDS, true); }
export function readOverlayGroupLenTime(view: DataView): number { return view.getFloat32(OVERLAY_COL_GROUP_LEN_TIME, true); }
export function readOverlayGroupLenInput(view: DataView): number { return view.getFloat32(OVERLAY_COL_GROUP_LEN_INPUT, true); }
export function readOverlayGroupLenGate(view: DataView): number { return view.getFloat32(OVERLAY_COL_GROUP_LEN_GATE, true); }
export function readOverlaySpeed(view: DataView): number { return view.getFloat32(OVERLAY_COL_SPEED, true); }

// ── Scene block ──────────────────────────────────────────────
export const SCENE_COL_CX                        = 0; // f32
export const SCENE_COL_CY                        = 4; // f32
export const SCENE_COL_CZ                        = 8; // f32
export const SCENE_COL_RADIUS                    = 12; // f32
export const SCENE_STRIDE                        = 16;

export function readSceneCX(view: DataView): number { return view.getFloat32(SCENE_COL_CX, true); }
export function readSceneCY(view: DataView): number { return view.getFloat32(SCENE_COL_CY, true); }
export function readSceneCZ(view: DataView): number { return view.getFloat32(SCENE_COL_CZ, true); }
export function readSceneRadius(view: DataView): number { return view.getFloat32(SCENE_COL_RADIUS, true); }

