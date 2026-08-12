import { postGoRecord } from "../../../../vscode-api";
import { encodeSceneCreate } from "../../../../../schema/input/input-encode-scene-tilt";

export const NODE_PALETTE_KIND_MIME = "application/x-wirefold-kind";

export function dropKindFromEvent(e: DragEvent): number | null {
  const raw = e.dataTransfer?.getData(NODE_PALETTE_KIND_MIME);
  if (!raw) return null;
  const id = Number(raw);
  return Number.isInteger(id) && id >= 0 ? id : null;
}

export function fireCreateAt(kindId: number, ndcX: number, ndcY: number): void {
  postGoRecord(encodeSceneCreate(kindId, ndcX, ndcY));
}
