import { postGoRecord } from "../src/webview/vscode-api";
import { encodePanelsToggle } from "../src/schema/input/input-encode";
import type { PanelFlag } from "../src/messages";
import { postLog } from "../src/webview/log/post";
import { usePanelFlags } from "../src/webview/three/controls/flags/panel-flags";

export function firePanelToggle(flag: PanelFlag, wasOpen: boolean) {
  postLog("panel-toggle-click", { flag, wasOpen });
  postGoRecord(encodePanelsToggle(flag));
}

export function usePanelOpen(flag: PanelFlag): boolean {
  const vals = usePanelFlags();
  return panelOpen(vals, flag);
}

export function panelOpen(vals: ReturnType<typeof usePanelFlags>, flag: PanelFlag): boolean {
  return vals ? (vals[flag] ?? false) : false;
}
