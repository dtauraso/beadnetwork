import { PANEL_FLAG_ORDER, type PanelFlag } from "./flags";
import { makeLeafValues } from "../../../valuefile/leaf-values";

export type PanelFlagVals = Record<PanelFlag, boolean>;

const values = makeLeafValues<PanelFlag>("Chrome/Panels/Panel/paths", PANEL_FLAG_ORDER);

const vals = Object.fromEntries(PANEL_FLAG_ORDER.map((f) => [f, false])) as PanelFlagVals;

export function startPanelReads(): void {
  values.u8(PANEL_FLAG_ORDER[0]);
}

export function panelFlagVals(): PanelFlagVals {
  for (const flag of PANEL_FLAG_ORDER) vals[flag] = values.u8(flag) !== 0;
  return vals;
}
