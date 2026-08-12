// OVERLAY_FLAGS_START
const OVERLAY_FLAG_NAMES = [
  "tori",
  "scenePoles",
  "nodePoles",
  "selSpherePoles",
  "handholds",
  "labelsGlobal",
  "overlays",

  "nodeBody",
  "nodeRing",
  "ringPick",
  "selectionRing",
  "hoverRing",
  "reachSphere",
] as const;
// OVERLAY_FLAGS_END

export type OverlayFlag = (typeof OVERLAY_FLAG_NAMES)[number];

export const OVERLAY_FLAG_ORDER = OVERLAY_FLAG_NAMES;

// EDIT_MSG_START

type EditMsg =

  | { type: "edit"; op: "update"; kind: "overlays"; attr: "toggle"; flag: OverlayFlag }
  | { type: "edit"; op: "update"; kind: "clock"; attr: "speed"; value: number }

  | { type: "edit"; op: "update"; kind: "distanceGroup"; attr: "length"; group: number; dir: "up" | "down" }

  | { type: "edit"; op: "update"; kind: "tiltVector"; attr: "theta" | "phi"; row: number; dir: "up" | "down" }

  | { type: "edit"; op: "update"; kind: "tiltVector"; attr: "reset"; row: number }

  | { type: "edit"; op: "update"; kind: "tiltVector"; attr: "start"; row: number }

  | { type: "edit"; op: "update"; kind: "scene"; attr: "selected"; tab: number }

  | { type: "edit"; op: "update"; kind: "scene"; attr: "latticePoints"; points: number }

  | { type: "edit"; op: "update"; kind: "scene"; attr: "create"; kindId: number; ndcX: number; ndcY: number }
  | { type: "edit"; op: "update"; kind: "scene"; attr: "delete"; row: number };
// EDIT_MSG_END

// RAW_INPUT_START

export type RawPointerKind = "pointerdown" | "pointermove" | "pointerup" | "wheel" | "home";

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
  fov: number; 
  hit: RawHit;
};
// RAW_INPUT_END

export type WebviewToHostMsg =
  | { type: "ready" }

  | { type: "go-record"; record: ArrayBuffer }
  | { type: "raw-input"; event: RawInputEvent }

  | { type: "save" }
  | { type: "webview-log"; entry: string }
  | EditMsg;

export type HostToWebviewMsg =

  | { type: "buffer-snapshot"; buffer: ArrayBuffer; tag: number; row?: number; gen: number };

export const WEBVIEW_TO_HOST_TYPES: ReadonlySet<WebviewToHostMsg["type"]> = new Set([
  "ready", "webview-log", "edit", "save", "raw-input", "go-record",
]);

const HOST_TO_WEBVIEW_TYPES: ReadonlySet<HostToWebviewMsg["type"]> = new Set([
  "buffer-snapshot",
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

export function parseHostToWebview(raw: unknown): HostToWebviewMsg | undefined {
  if (!raw || typeof raw !== "object") return undefined;
  const m = raw as Record<string, unknown>;
  const t = m.type;
  if (typeof t !== "string" || !HOST_TO_WEBVIEW_TYPES.has(t as HostToWebviewMsg["type"])) {
    return undefined;
  }
  switch (t) {
    case "buffer-snapshot":

      return m.buffer instanceof ArrayBuffer && typeof m.tag === "number"
        ? (m as unknown as HostToWebviewMsg)
        : undefined;
    default:
      return undefined;
  }
}
