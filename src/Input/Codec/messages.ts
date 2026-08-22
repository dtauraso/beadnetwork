import type { OverlayFlag } from "../../Overlay/flags";
import type { PanelFlag } from "../../Chrome/Panels/Panel/flags";

export type { OverlayFlag, PanelFlag };

// EDIT_MSG_START

type EditMsg =

  | { type: "edit"; op: "update"; kind: "overlays"; attr: "toggle"; flag: OverlayFlag }
  | { type: "edit"; op: "update"; kind: "panels"; attr: "toggle"; flag: PanelFlag }
  | { type: "edit"; op: "update"; kind: "clock"; attr: "speed"; value: number }

  | { type: "edit"; op: "update"; kind: "tiltVector"; attr: "phi"; row: number; dir: "up" | "down" }

  | { type: "edit"; op: "update"; kind: "tiltVector"; attr: "reset"; row: number }

  | { type: "edit"; op: "update"; kind: "tiltVector"; attr: "start"; row: number }

  | { type: "edit"; op: "update"; kind: "scene"; attr: "selected"; tab: number }

  | { type: "edit"; op: "update"; kind: "scene"; attr: "latticePoints"; points: number }

  | { type: "edit"; op: "update"; kind: "scene"; attr: "create"; kindId: number; ndcX: number; ndcY: number }
  | { type: "edit"; op: "update"; kind: "scene"; attr: "delete"; row: number }

  | { type: "edit"; op: "update"; kind: "node"; attr: "dragPhi"; row: number }
  | { type: "edit"; op: "update"; kind: "node"; attr: "dragMaxTheta"; row: number; piMultiple: number }
  | { type: "edit"; op: "update"; kind: "node"; attr: "dragActive"; row: number }
  | { type: "edit"; op: "update"; kind: "node"; attr: "kindActive"; row: number }
  | { type: "edit"; op: "update"; kind: "node"; attr: "selfDragPhi"; row: number }
  | { type: "edit"; op: "update"; kind: "node"; attr: "selfDragMaxTheta"; row: number; piMultiple: number }
  | { type: "edit"; op: "update"; kind: "node"; attr: "selfDragActive"; row: number }
  | { type: "edit"; op: "update"; kind: "node"; attr: "dragR"; row: number }
  | { type: "edit"; op: "update"; kind: "node"; attr: "selfDragR"; row: number }
  | { type: "edit"; op: "update"; kind: "edge"; attr: "dragActive"; row: number };
// EDIT_MSG_END

// RAW_INPUT_START

export type RawPointerKind = "pointerdown" | "pointermove" | "pointerup" | "wheel" | "home" | "delete" | "key";

export type RawHit = {
  kind: "port" | "handhold" | "node" | "edge" | "torus" | "empty";
  isInput: boolean;

  nodeRow: number;
  portRow: number;

  edgeRow: number;
};

export type RawInputEvent = {
  kind: RawPointerKind;
  x: number; 
  y: number; 
  rectLeft: number;
  rectTop: number;
  rectWidth: number;
  rectHeight: number;
  button: number; 
  ctrl: boolean;
  shift: boolean;
  alt: boolean;
  meta: boolean;
  deltaX: number; 
  deltaY: number;
  hit: RawHit;
  key?: string;
};
// RAW_INPUT_END

export type WebviewToHostMsg =
  | { type: "ready" }

  | { type: "go-record"; record: ArrayBuffer }
  | { type: "raw-input"; event: RawInputEvent }

  | { type: "save" }
  | { type: "webview-log"; entry: string }
  | EditMsg;

export const WEBVIEW_TO_HOST_TYPES: ReadonlySet<WebviewToHostMsg["type"]> = new Set([
  "ready", "webview-log", "edit", "save", "raw-input", "go-record",
]);

export function parseWebviewToHost(raw: unknown): WebviewToHostMsg | undefined {
  if (!raw || typeof raw !== "object") return undefined;
  const t = (raw as { type?: unknown }).type;
  if (typeof t !== "string" || !WEBVIEW_TO_HOST_TYPES.has(t as WebviewToHostMsg["type"])) {
    return undefined;
  }
  const m = raw as Record<string, unknown>;
  switch (t) {
    case "webview-log":
      return typeof m.entry === "string" ? (m as unknown as WebviewToHostMsg) : undefined;
    case "go-record":

      return m.record instanceof ArrayBuffer ? (m as unknown as WebviewToHostMsg) : undefined;
    default:
      return m as unknown as WebviewToHostMsg;
  }
}
