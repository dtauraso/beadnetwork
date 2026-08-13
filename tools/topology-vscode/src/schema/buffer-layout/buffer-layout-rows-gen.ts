// ── Node block ───────────────────────────────────────────────
export const NODE_COL_NODE_ID                    = 0; // i32
export const NODE_COL_CX                         = 4; // f32
export const NODE_COL_CY                         = 8; // f32
export const NODE_COL_CZ                         = 12; // f32
export const NODE_COL_RADIUS                     = 16; // f32
export const NODE_COL_SPHERE_R                   = 20; // f32
export const NODE_COL_VRX                        = 24; // f32
export const NODE_COL_VRY                        = 28; // f32
export const NODE_COL_VRZ                        = 32; // f32
export const NODE_COL_FRX                        = 36; // f32
export const NODE_COL_FRY                        = 40; // f32
export const NODE_COL_FRZ                        = 44; // f32
export const NODE_COL_POLE_PHI                   = 48; // f32
export const NODE_COL_POLE_THETA                 = 52; // f32
export const NODE_COL_RING_AXIS_PHI              = 56; // f32
export const NODE_COL_RING_AXIS_THETA            = 60; // f32
export const NODE_COL_TOP_TILT_VECTOR_LEN        = 64; // f32
export const NODE_COL_TOP_TILT_VECTOR_THETA      = 68; // f32
export const NODE_COL_BOTTOM_TILT_VECTOR_THETA   = 72; // f32
export const NODE_COL_COPLANAR_NORMAL_THETA      = 76; // f32
export const NODE_COL_RECEIVED_VECTOR_LEN        = 80; // f32
export const NODE_COL_RECEIVED_VECTOR_THETA      = 84; // f32
export const NODE_COL_SELECTED                   = 88; // u8
export const NODE_COL_KIND_ID                    = 89; // u8
export const NODE_COL_LABEL_OFF                  = 90; // u32
export const NODE_COL_LABEL_LEN                  = 94; // u32
export const NODE_COL_HOVERED                    = 98; // u8
export const NODE_COL_LATCHED_SEL                = 99; // u8
export const NODE_COL_LATTICE_POINTS             = 100; // u8
export const NODE_COL_ROUNDS_TO_PARALLEL         = 101; // i32
export const NODE_COL_MSGS_TO_PARALLEL           = 105; // i32
export const NODE_STRIDE                         = 109;

export function readNodeNodeId(view: DataView, row: number): number { return view.getInt32(row * NODE_STRIDE + NODE_COL_NODE_ID, true); }
export function readNodeCX(view: DataView, row: number): number { return view.getFloat32(row * NODE_STRIDE + NODE_COL_CX, true); }
export function readNodeCY(view: DataView, row: number): number { return view.getFloat32(row * NODE_STRIDE + NODE_COL_CY, true); }
export function readNodeCZ(view: DataView, row: number): number { return view.getFloat32(row * NODE_STRIDE + NODE_COL_CZ, true); }
export function readNodeRadius(view: DataView, row: number): number { return view.getFloat32(row * NODE_STRIDE + NODE_COL_RADIUS, true); }
export function readNodeSphereR(view: DataView, row: number): number { return view.getFloat32(row * NODE_STRIDE + NODE_COL_SPHERE_R, true); }
export function readNodeVRX(view: DataView, row: number): number { return view.getFloat32(row * NODE_STRIDE + NODE_COL_VRX, true); }
export function readNodeVRY(view: DataView, row: number): number { return view.getFloat32(row * NODE_STRIDE + NODE_COL_VRY, true); }
export function readNodeVRZ(view: DataView, row: number): number { return view.getFloat32(row * NODE_STRIDE + NODE_COL_VRZ, true); }
export function readNodeFRX(view: DataView, row: number): number { return view.getFloat32(row * NODE_STRIDE + NODE_COL_FRX, true); }
export function readNodeFRY(view: DataView, row: number): number { return view.getFloat32(row * NODE_STRIDE + NODE_COL_FRY, true); }
export function readNodeFRZ(view: DataView, row: number): number { return view.getFloat32(row * NODE_STRIDE + NODE_COL_FRZ, true); }
export function readNodePolePhi(view: DataView, row: number): number { return view.getFloat32(row * NODE_STRIDE + NODE_COL_POLE_PHI, true); }
export function readNodePoleTheta(view: DataView, row: number): number { return view.getFloat32(row * NODE_STRIDE + NODE_COL_POLE_THETA, true); }
export function readNodeRingAxisPhi(view: DataView, row: number): number { return view.getFloat32(row * NODE_STRIDE + NODE_COL_RING_AXIS_PHI, true); }
export function readNodeRingAxisTheta(view: DataView, row: number): number { return view.getFloat32(row * NODE_STRIDE + NODE_COL_RING_AXIS_THETA, true); }
export function readNodeTopTiltVectorLen(view: DataView, row: number): number { return view.getFloat32(row * NODE_STRIDE + NODE_COL_TOP_TILT_VECTOR_LEN, true); }
export function readNodeTopTiltVectorTheta(view: DataView, row: number): number { return view.getFloat32(row * NODE_STRIDE + NODE_COL_TOP_TILT_VECTOR_THETA, true); }
export function readNodeBottomTiltVectorTheta(view: DataView, row: number): number { return view.getFloat32(row * NODE_STRIDE + NODE_COL_BOTTOM_TILT_VECTOR_THETA, true); }
export function readNodeCoplanarNormalTheta(view: DataView, row: number): number { return view.getFloat32(row * NODE_STRIDE + NODE_COL_COPLANAR_NORMAL_THETA, true); }
export function readNodeReceivedVectorLen(view: DataView, row: number): number { return view.getFloat32(row * NODE_STRIDE + NODE_COL_RECEIVED_VECTOR_LEN, true); }
export function readNodeReceivedVectorTheta(view: DataView, row: number): number { return view.getFloat32(row * NODE_STRIDE + NODE_COL_RECEIVED_VECTOR_THETA, true); }
export function readNodeSelected(view: DataView, row: number): number { return view.getUint8(row * NODE_STRIDE + NODE_COL_SELECTED); }
export function readNodeKindId(view: DataView, row: number): number { return view.getUint8(row * NODE_STRIDE + NODE_COL_KIND_ID); }
export function readNodeLabelOff(view: DataView, row: number): number { return view.getUint32(row * NODE_STRIDE + NODE_COL_LABEL_OFF, true); }
export function readNodeLabelLen(view: DataView, row: number): number { return view.getUint32(row * NODE_STRIDE + NODE_COL_LABEL_LEN, true); }
export function readNodeHovered(view: DataView, row: number): number { return view.getUint8(row * NODE_STRIDE + NODE_COL_HOVERED); }
export function readNodeLatchedSel(view: DataView, row: number): number { return view.getUint8(row * NODE_STRIDE + NODE_COL_LATCHED_SEL); }
export function readNodeLatticePoints(view: DataView, row: number): number { return view.getUint8(row * NODE_STRIDE + NODE_COL_LATTICE_POINTS); }
export function readNodeRoundsToParallel(view: DataView, row: number): number { return view.getInt32(row * NODE_STRIDE + NODE_COL_ROUNDS_TO_PARALLEL, true); }
export function readNodeMsgsToParallel(view: DataView, row: number): number { return view.getInt32(row * NODE_STRIDE + NODE_COL_MSGS_TO_PARALLEL, true); }

// ── Interior block ───────────────────────────────────────────
export const INTERIOR_COL_PRESENT                = 0; // u8
export const INTERIOR_COL_VALUE                  = 1; // i32
export const INTERIOR_COL_OX                     = 5; // f32
export const INTERIOR_COL_OY                     = 9; // f32
export const INTERIOR_COL_OZ                     = 13; // f32
export const INTERIOR_STRIDE                     = 17;

export function readInteriorPresent(view: DataView, row: number): number { return view.getUint8(row * INTERIOR_STRIDE + INTERIOR_COL_PRESENT); }
export function readInteriorValue(view: DataView, row: number): number { return view.getInt32(row * INTERIOR_STRIDE + INTERIOR_COL_VALUE, true); }
export function readInteriorOX(view: DataView, row: number): number { return view.getFloat32(row * INTERIOR_STRIDE + INTERIOR_COL_OX, true); }
export function readInteriorOY(view: DataView, row: number): number { return view.getFloat32(row * INTERIOR_STRIDE + INTERIOR_COL_OY, true); }
export function readInteriorOZ(view: DataView, row: number): number { return view.getFloat32(row * INTERIOR_STRIDE + INTERIOR_COL_OZ, true); }

