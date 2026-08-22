import type { PanelFlag } from "./flags";

// EDIT_MSG_START
export type PanelEditMsg =
  | { type: "edit"; op: "update"; kind: "panels"; attr: "toggle"; flag: PanelFlag };
// EDIT_MSG_END
