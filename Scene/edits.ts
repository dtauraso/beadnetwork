// EDIT_MSG_START
export type SceneEditMsg =
  | { type: "edit"; op: "update"; kind: "scene"; attr: "selected"; tab: number }
  | { type: "edit"; op: "update"; kind: "scene"; attr: "latticePoints"; points: number }
  | { type: "edit"; op: "update"; kind: "scene"; attr: "create"; kindId: number; ndcX: number; ndcY: number }
  | { type: "edit"; op: "update"; kind: "scene"; attr: "delete"; row: number }
  | { type: "edit"; op: "update"; kind: "scene"; attr: "viewport"; width: number; height: number }
  | { type: "edit"; op: "update"; kind: "tiltVector"; attr: "phi"; row: number; dir: "up" | "down" }
  | { type: "edit"; op: "update"; kind: "tiltVector"; attr: "reset"; row: number }
  | { type: "edit"; op: "update"; kind: "tiltVector"; attr: "start"; row: number };
// EDIT_MSG_END
