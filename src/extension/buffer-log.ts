import { decodeViewFrame } from "../webview/decode/buffer-decode-view";
import { decodeEventLines } from "../webview/decode/decode-event-line";
import type { DecodedEvents } from "../webview/decode/buffer-decode-shared";

export type DecodedEventLine =
  | { step: number; kind: "recv" | "fire"; node: string; port?: string; value?: number }
  | { step: number; kind: "send"; node: string; port?: string; value?: number; beadSteps?: number; target?: string; targetHandle?: string }
  | { step: number; kind: "edge-bead"; node: string; port: string; value?: number; x: number; y: number; z: number; f: number; bead?: number }
  | { step: number; kind: "geometry"; edge: string; sx: number; sy: number; sz: number; ex: number; ey: number; ez: number }
  | { step: number; kind: "arrive"; node: string; port: string; value?: number; bead?: number }
  | { step: number; kind: "node-geometry"; node: string; label?: string; nodeKind?: string; nx: number; ny: number; nz: number; radius: number; sphereR?: number; ports: { name: string; isInput: boolean; px: number; py: number; pz: number; dx: number; dy: number; dz: number }[] }
  | { step: number; kind: "node-bead"; node: string; row: number; col: number; present: boolean; value: number; x: number; y: number; z: number }
  | { step: number; kind: "camera"; px: number; py: number; pz: number; r: number; posTheta: number; posPhi: number; upTheta: number; upPhi: number }
  | { step: number; kind: "scene-sphere"; cx: number; cy: number; cz: number; radius: number }
  | { step: number; kind: "scene-tori"; visible: boolean }
  | { step: number; kind: "scene-poles"; visible: boolean }
  | { step: number; kind: "node-poles"; visible: boolean }
  | { step: number; kind: "handholds"; visible: boolean }
  | { step: number; kind: "labels-global"; visible: boolean }
  | { step: number; kind: "overlays-vis"; visible: boolean }

  | { step: number; kind: "select"; node: string }
  | { step: number; kind: "hover"; node: string; port?: string; value?: number }

  | {
      step: number; kind: "breadcrumb"; label: string; debug: boolean;
      node: string; port?: string; value?: number; x: number; y: number; z: number;
      nodeRow: number; portRow: number; targetRow: number; targetPortRow: number; edgeRow: number; slot: number;
      target?: string; text?: string;
    };

export function decodeBufferLog(viewFrameBuf: ArrayBuffer, breadcrumbsOnly = false): string {
  const dv = decodeViewFrame(viewFrameBuf);
  if (!dv) return "";
  return decodeStreamFrameEvents(dv.events, breadcrumbsOnly);
}

export function decodeStreamFrameEvents(events: DecodedEvents, breadcrumbsOnly = false): string {
  const now = Date.now();
  let out = "";
  for (const line of decodeEventLines(events, breadcrumbsOnly)) {
    out += JSON.stringify({ ts_ms: now, src: "go", ...line }) + "\n";
  }
  return out;
}
