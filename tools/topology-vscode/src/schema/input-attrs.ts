// input-attrs.ts — edit-update ATTR index constants, hand-authored (not generated: only
// the fingerprint/kind-bytes/enum-orderings come from input-layout-gen.ts). Must match
// input_fingerprint.go's numbering; input-encode.ts and input-decode.ts both read these.

export const IN_OVERLAY_ATTR_TOGGLE = 0;
export const IN_CLOCK_ATTR_SPEED = 1;
export const IN_DISTANCE_GROUP_ATTR_LENGTH = 2;
export const IN_SCENE_ATTR_SELECTED = 3;
export const IN_TILT_VECTOR_ATTR_THETA = 4;
// attr 5 (phi) is a GAP — the tilt vector is θ-only end to end now (task/drop-tilt-vector-phi);
// never renumber the survivors.
export const IN_TILT_VECTOR_ATTR_RESET = 6;
export const IN_TILT_VECTOR_ATTR_START = 7;
export const IN_SCENE_ATTR_LATTICE_POINTS = 8;
export const IN_SCENE_ATTR_CREATE = 9;
export const IN_SCENE_ATTR_DELETE = 10;
