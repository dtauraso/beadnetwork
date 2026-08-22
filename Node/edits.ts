// EDIT_MSG_START
export type NodeEditMsg =
  | { type: "edit"; op: "update"; kind: "node"; attr: "dragPhi"; row: number }
  | { type: "edit"; op: "update"; kind: "node"; attr: "dragMaxTheta"; row: number; piMultiple: number }
  | { type: "edit"; op: "update"; kind: "node"; attr: "dragActive"; row: number }
  | { type: "edit"; op: "update"; kind: "node"; attr: "kindActive"; row: number }
  | { type: "edit"; op: "update"; kind: "node"; attr: "selfDragPhi"; row: number }
  | { type: "edit"; op: "update"; kind: "node"; attr: "selfDragMaxTheta"; row: number; piMultiple: number }
  | { type: "edit"; op: "update"; kind: "node"; attr: "selfDragActive"; row: number }
  | { type: "edit"; op: "update"; kind: "node"; attr: "dragR"; row: number }
  | { type: "edit"; op: "update"; kind: "node"; attr: "selfDragR"; row: number };
// EDIT_MSG_END
