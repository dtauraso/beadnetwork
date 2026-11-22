// EDIT_MSG_START
export type ClockEditMsg =
  | { type: "edit"; op: "update"; kind: "clock"; attr: "speed"; value: number };
// EDIT_MSG_END
