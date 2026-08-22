import { editUpdate } from "../Input/Codec/attr-index";

function nodeRowEdit(attr: string, nodeRow: number): ArrayBuffer {
  const w = editUpdate("node", attr);
  w.u8(nodeRow);
  return w.toArrayBuffer();
}

function nodeMaxThetaEdit(attr: string, nodeRow: number, piMultiple: number): ArrayBuffer {
  const w = editUpdate("node", attr);
  w.u8(nodeRow);
  w.f32(piMultiple);
  return w.toArrayBuffer();
}

export const encodeNodeDragPhiToggle = (row: number) => nodeRowEdit("dragPhi", row);
export const encodeNodeDragActiveToggle = (row: number) => nodeRowEdit("dragActive", row);
export const encodeNodeKindActiveToggle = (row: number) => nodeRowEdit("kindActive", row);
export const encodeNodeDragRToggle = (row: number) => nodeRowEdit("dragR", row);
export const encodeNodeSelfDragRToggle = (row: number) => nodeRowEdit("selfDragR", row);
export const encodeNodeSelfDragPhiToggle = (row: number) => nodeRowEdit("selfDragPhi", row);
export const encodeNodeSelfDragActiveToggle = (row: number) => nodeRowEdit("selfDragActive", row);

export const encodeNodeDragMaxTheta = (row: number, piMultiple: number) =>
  nodeMaxThetaEdit("dragMaxTheta", row, piMultiple);
export const encodeNodeSelfDragMaxTheta = (row: number, piMultiple: number) =>
  nodeMaxThetaEdit("selfDragMaxTheta", row, piMultiple);
