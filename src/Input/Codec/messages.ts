import type { OverlayEditMsg } from "../../Overlay/edits";
import type { PanelEditMsg } from "../../Chrome/Panels/Panel/edits";
import type { ClockEditMsg } from "../../Clock/edits";
import type { SceneEditMsg } from "../../Scene/edits";
import type { NodeEditMsg } from "../../Node/edits";
import type { EdgeEditMsg } from "../../Node/Edge/edits";
import type { OverlayFlag } from "../../Overlay/flags";
import type { PanelFlag } from "../../Chrome/Panels/Panel/flags";

export type { OverlayFlag, PanelFlag };

type EditMsg =
  | OverlayEditMsg
  | PanelEditMsg
  | ClockEditMsg
  | SceneEditMsg
  | NodeEditMsg
  | EdgeEditMsg;

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
