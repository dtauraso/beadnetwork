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
export const NODE_COL_POLE_THETA                 = 48; // f32
export const NODE_COL_POLE_PHI                   = 52; // f32
export const NODE_COL_RING_AXIS_THETA            = 56; // f32
export const NODE_COL_RING_AXIS_PHI              = 60; // f32
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
export function readNodePoleTheta(view: DataView, row: number): number { return view.getFloat32(row * NODE_STRIDE + NODE_COL_POLE_THETA, true); }
export function readNodePolePhi(view: DataView, row: number): number { return view.getFloat32(row * NODE_STRIDE + NODE_COL_POLE_PHI, true); }
export function readNodeRingAxisTheta(view: DataView, row: number): number { return view.getFloat32(row * NODE_STRIDE + NODE_COL_RING_AXIS_THETA, true); }
export function readNodeRingAxisPhi(view: DataView, row: number): number { return view.getFloat32(row * NODE_STRIDE + NODE_COL_RING_AXIS_PHI, true); }
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

// ── ChainBead block ──────────────────────────────────────────
export const CHAIN_BEAD_COL_OX                   = 0; // f32
export const CHAIN_BEAD_COL_OY                   = 4; // f32
export const CHAIN_BEAD_COL_OZ                   = 8; // f32
export const CHAIN_BEAD_COL_LIT                  = 12; // u8
export const CHAIN_BEAD_COL_LIT_VALUE            = 13; // i32
export const CHAIN_BEAD_STRIDE                   = 17;

export function readChainBeadOX(view: DataView, row: number): number { return view.getFloat32(row * CHAIN_BEAD_STRIDE + CHAIN_BEAD_COL_OX, true); }
export function readChainBeadOY(view: DataView, row: number): number { return view.getFloat32(row * CHAIN_BEAD_STRIDE + CHAIN_BEAD_COL_OY, true); }
export function readChainBeadOZ(view: DataView, row: number): number { return view.getFloat32(row * CHAIN_BEAD_STRIDE + CHAIN_BEAD_COL_OZ, true); }
export function readChainBeadLit(view: DataView, row: number): number { return view.getUint8(row * CHAIN_BEAD_STRIDE + CHAIN_BEAD_COL_LIT); }
export function readChainBeadLitValue(view: DataView, row: number): number { return view.getInt32(row * CHAIN_BEAD_STRIDE + CHAIN_BEAD_COL_LIT_VALUE, true); }

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

// ── Edge block ───────────────────────────────────────────────
export const EDGE_COL_SX                         = 0; // f32
export const EDGE_COL_SY                         = 4; // f32
export const EDGE_COL_SZ                         = 8; // f32
export const EDGE_COL_EX                         = 12; // f32
export const EDGE_COL_EY                         = 16; // f32
export const EDGE_COL_EZ                         = 20; // f32
export const EDGE_COL_EDGE_LABEL_OFF             = 24; // u32
export const EDGE_COL_EDGE_LABEL_LEN             = 28; // u32
export const EDGE_STRIDE                         = 32;

export function readEdgeSX(view: DataView, row: number): number { return view.getFloat32(row * EDGE_STRIDE + EDGE_COL_SX, true); }
export function readEdgeSY(view: DataView, row: number): number { return view.getFloat32(row * EDGE_STRIDE + EDGE_COL_SY, true); }
export function readEdgeSZ(view: DataView, row: number): number { return view.getFloat32(row * EDGE_STRIDE + EDGE_COL_SZ, true); }
export function readEdgeEX(view: DataView, row: number): number { return view.getFloat32(row * EDGE_STRIDE + EDGE_COL_EX, true); }
export function readEdgeEY(view: DataView, row: number): number { return view.getFloat32(row * EDGE_STRIDE + EDGE_COL_EY, true); }
export function readEdgeEZ(view: DataView, row: number): number { return view.getFloat32(row * EDGE_STRIDE + EDGE_COL_EZ, true); }
export function readEdgeEdgeLabelOff(view: DataView, row: number): number { return view.getUint32(row * EDGE_STRIDE + EDGE_COL_EDGE_LABEL_OFF, true); }
export function readEdgeEdgeLabelLen(view: DataView, row: number): number { return view.getUint32(row * EDGE_STRIDE + EDGE_COL_EDGE_LABEL_LEN, true); }

// ── Event block ──────────────────────────────────────────────
export const EVENT_COL_KIND                      = 0; // u8
export const EVENT_COL_NODE_ROW                  = 1; // i32
export const EVENT_COL_PORT_ROW                  = 5; // i32
export const EVENT_COL_TARGET_ROW                = 9; // i32
export const EVENT_COL_TARGET_PORT_ROW           = 13; // i32
export const EVENT_COL_EDGE_ROW                  = 17; // i32
export const EVENT_COL_SLOT                      = 21; // i32
export const EVENT_COL_VALUE                     = 25; // i32
export const EVENT_COL_BEAD                      = 29; // u32
export const EVENT_COL_BEAD_STEPS                = 33; // f32
export const EVENT_COL_SIM_LATENCY_MS            = 37; // f32
export const EVENT_COL_X                         = 41; // f32
export const EVENT_COL_Y                         = 45; // f32
export const EVENT_COL_Z                         = 49; // f32
export const EVENT_COL_F                         = 53; // f32
export const EVENT_COL_LABEL                     = 57; // u8
export const EVENT_COL_DEBUG                     = 58; // u8
export const EVENT_COL_TEXT_OFF                  = 59; // u32
export const EVENT_COL_TEXT_LEN                  = 63; // u32
export const EVENT_STRIDE                        = 67;

export function readEventKind(view: DataView, row: number): number { return view.getUint8(row * EVENT_STRIDE + EVENT_COL_KIND); }
export function readEventNodeRow(view: DataView, row: number): number { return view.getInt32(row * EVENT_STRIDE + EVENT_COL_NODE_ROW, true); }
export function readEventPortRow(view: DataView, row: number): number { return view.getInt32(row * EVENT_STRIDE + EVENT_COL_PORT_ROW, true); }
export function readEventTargetRow(view: DataView, row: number): number { return view.getInt32(row * EVENT_STRIDE + EVENT_COL_TARGET_ROW, true); }
export function readEventTargetPortRow(view: DataView, row: number): number { return view.getInt32(row * EVENT_STRIDE + EVENT_COL_TARGET_PORT_ROW, true); }
export function readEventEdgeRow(view: DataView, row: number): number { return view.getInt32(row * EVENT_STRIDE + EVENT_COL_EDGE_ROW, true); }
export function readEventSlot(view: DataView, row: number): number { return view.getInt32(row * EVENT_STRIDE + EVENT_COL_SLOT, true); }
export function readEventValue(view: DataView, row: number): number { return view.getInt32(row * EVENT_STRIDE + EVENT_COL_VALUE, true); }
export function readEventBead(view: DataView, row: number): number { return view.getUint32(row * EVENT_STRIDE + EVENT_COL_BEAD, true); }
export function readEventBeadSteps(view: DataView, row: number): number { return view.getFloat32(row * EVENT_STRIDE + EVENT_COL_BEAD_STEPS, true); }
export function readEventSimLatencyMs(view: DataView, row: number): number { return view.getFloat32(row * EVENT_STRIDE + EVENT_COL_SIM_LATENCY_MS, true); }
export function readEventX(view: DataView, row: number): number { return view.getFloat32(row * EVENT_STRIDE + EVENT_COL_X, true); }
export function readEventY(view: DataView, row: number): number { return view.getFloat32(row * EVENT_STRIDE + EVENT_COL_Y, true); }
export function readEventZ(view: DataView, row: number): number { return view.getFloat32(row * EVENT_STRIDE + EVENT_COL_Z, true); }
export function readEventF(view: DataView, row: number): number { return view.getFloat32(row * EVENT_STRIDE + EVENT_COL_F, true); }
export function readEventLabel(view: DataView, row: number): number { return view.getUint8(row * EVENT_STRIDE + EVENT_COL_LABEL); }
export function readEventDebug(view: DataView, row: number): number { return view.getUint8(row * EVENT_STRIDE + EVENT_COL_DEBUG); }
export function readEventTextOff(view: DataView, row: number): number { return view.getUint32(row * EVENT_STRIDE + EVENT_COL_TEXT_OFF, true); }
export function readEventTextLen(view: DataView, row: number): number { return view.getUint32(row * EVENT_STRIDE + EVENT_COL_TEXT_LEN, true); }

