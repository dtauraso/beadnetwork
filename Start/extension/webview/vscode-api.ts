import type { WebviewToHostMsg } from "./messages";
import { encodeRawInput } from "../../../Categories/Input/Drag/encode";
import type { RawInputEvent } from "../../../Categories/Input/Drag/raw-input";

declare function acquireVsCodeApi(): {
  postMessage(msg: WebviewToHostMsg): void;
  setState(s: unknown): void;
  getState(): unknown;
};

type VsCodeApi = ReturnType<typeof acquireVsCodeApi>;
const w = window as unknown as { __vscodeApi?: VsCodeApi };
export const vscode: VsCodeApi = w.__vscodeApi ?? (w.__vscodeApi = acquireVsCodeApi());

export function postGoRecord(record: ArrayBuffer): void {
  vscode.postMessage({ type: "go-record", record });
}

export function sendRawInput(event: RawInputEvent): void {
  postGoRecord(encodeRawInput(event));
}
