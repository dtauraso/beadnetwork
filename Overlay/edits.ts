import type { OverlayFlag } from "./flags";

// EDIT_MSG_START
export type OverlayEditMsg =
  | { type: "edit"; op: "update"; kind: "overlays"; attr: "toggle"; flag: OverlayFlag };
// EDIT_MSG_END
