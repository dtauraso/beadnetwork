import { PANEL_FLAG_ORDER, type PanelFlag } from "./flags";
import { panelFlagVals, startPanelReads, type PanelFlagVals } from "./panel-leaves";

export type { PanelFlagVals };

export function readPanelFlags(): PanelFlagVals {
  startPanelReads();
  return panelFlagVals();
}

export function panelFlag(name: PanelFlag): boolean {
  return readPanelFlags()[name];
}

export function panelFlagSignature(): string {
  const vals = readPanelFlags();
  return PANEL_FLAG_ORDER.map((flag) => (vals[flag] ? "1" : "0")).join("");
}
