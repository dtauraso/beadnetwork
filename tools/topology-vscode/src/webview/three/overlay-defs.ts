// overlay-defs.ts — the config table for the overlay toggle buttons and the groups the
// popover clusters them into. Data only; the rows/groups themselves render elsewhere.

import type { ToggleCfg } from "./overlay-toggle";

export const guidelinesCfg: ToggleCfg = {
  flag: "overlays",
  default: true,
  active: (v) => v,
  icon: "▦",
  label: "overlays",
  title: (a) => (a ? "Hide overlays" : "Show overlays"),
  payload: (v) => ({ flag: "overlays", was: v }),
};

const ringsCfg: ToggleCfg = {
  flag: "tori",
  default: true,
  active: (v) => v,
  icon: "◎",
  label: "rings",
  title: (a) => (a ? "Hide polar rings" : "Show polar rings"),
  payload: (v) => ({ flag: "tori", was: v }),
};

const scenePolesCfg: ToggleCfg = {
  flag: "scenePoles",
  default: true,
  active: (v) => v,
  icon: "⊹",
  label: "scene poles",
  title: (a) => (a ? "Hide scene pole frame" : "Show scene pole frame"),
  payload: (v) => ({ flag: "scenePoles", was: v }),
};

const nodePolesCfg: ToggleCfg = {
  flag: "nodePoles",
  default: true,
  active: (v) => v,
  icon: "⊹",
  label: "node poles",
  title: (a) => (a ? "Hide node pole frames" : "Show node pole frames"),
  payload: (v) => ({ flag: "nodePoles", was: v }),
};

const selSpherePolesCfg: ToggleCfg = {
  flag: "selSpherePoles",
  default: true,
  active: (v) => v,
  // Was `select ⬡` — the one row whose glyph trailed its words. It leads now, like the rest.
  icon: "⬡",
  label: "select",
  title: (a) => (a ? "Hide select-sphere poles" : "Show select-sphere poles"),
  payload: (v) => ({ flag: "selSpherePoles", was: v }),
};

const handholdsCfg: ToggleCfg = {
  flag: "handholds",
  default: true,
  active: (v) => v !== false,
  icon: "⊙",
  label: "grips",
  title: (a) => (a ? "Hide rotation grips" : "Show rotation grips"),
  payload: (v) => ({ flag: "handholds", was: v }),
};

const globalLabelsCfg: ToggleCfg = {
  flag: "labelsGlobal",
  default: false,
  active: (v) => !v,
  icon: (v) => (v ? "▴" : "▾"),
  label: "labels",
  title: (a) => (a ? "Hide labels" : "Show labels"),
  payload: (v) => ({ flag: "labelsGlobal", wasHidden: v }),
};

// The NODE-LOCAL drawings. Everything above this line is scene furniture drawn AROUND the
// nodes; these six are the node itself, and until now nothing could turn any of them off.
// Each is `active: (v) => v` and default-on: the flag says "drawn", and a node with all six
// off simply is not drawn.

const nodeBodyCfg: ToggleCfg = {
  flag: "nodeBody",
  default: true,
  active: (v) => v,
  icon: "●",
  label: "body",
  title: (a) => (a ? "Hide node bodies" : "Show node bodies"),
  payload: (v) => ({ flag: "nodeBody", was: v }),
};

const nodeRingCfg: ToggleCfg = {
  flag: "nodeRing",
  default: true,
  active: (v) => v,
  icon: "○",
  label: "ring",
  title: (a) => (a ? "Hide node rings" : "Show node rings"),
  payload: (v) => ({ flag: "nodeRing", was: v }),
};

const ringPickCfg: ToggleCfg = {
  flag: "ringPick",
  default: true,
  active: (v) => v,
  icon: "◌",
  // The band that takes a ring click, painted so you can see where it is. Like every other
  // overlay this shows or hides a drawing and nothing else — the ring takes clicks either
  // way (that is select mode's job), so this can never quietly disable an interaction.
  label: "ring band",
  title: (a) => (a ? "Hide the ring's click band" : "Show the ring's click band"),
  payload: (v) => ({ flag: "ringPick", was: v }),
};

const selectionRingCfg: ToggleCfg = {
  flag: "selectionRing",
  default: true,
  active: (v) => v,
  icon: "◉",
  label: "selection",
  title: (a) => (a ? "Hide the selection ring" : "Show the selection ring"),
  payload: (v) => ({ flag: "selectionRing", was: v }),
};

const hoverRingCfg: ToggleCfg = {
  flag: "hoverRing",
  default: true,
  active: (v) => v,
  icon: "◍",
  label: "hover",
  title: (a) => (a ? "Hide the hover ring" : "Show the hover ring"),
  payload: (v) => ({ flag: "hoverRing", was: v }),
};

const reachSphereCfg: ToggleCfg = {
  flag: "reachSphere",
  default: true,
  active: (v) => v,
  icon: "⌾",
  label: "reach sphere",
  title: (a) => (a ? "Hide the reach sphere" : "Show the reach sphere"),
  payload: (v) => ({ flag: "reachSphere", was: v }),
};

// `under` names the cfg a row NESTS beneath: that row renders indented and is disabled
// whenever its parent is off. View structure only — the gating that actually suppresses the
// drawing is Go-owned and lives in the renderer, so a disabled row here is never the only
// thing holding a child off.
export type OverlayGroup = { heading: string; cfgs: ToggleCfg[]; under?: Partial<Record<string, ToggleCfg>> };

// EVERY overlay sits in a cluster — there is no loose row at the top level of the popover,
// so the list reads as "which part of the picture" rather than as one flat inventory. The
// clusters answer that question in one word each: what a NODE is made of, what marks the
// node you are touching, the scene furniture you navigate BY, the pole frames, the text.
//
// NODE and STATE are the new ones. The split between them is what changes the drawing
// permanently (a node's body and ring are there whatever you do) versus what appears
// because of where the pointer or the selection is right now.
export const OVERLAY_GROUPS: OverlayGroup[] = [
  { heading: "NODE",   cfgs: [nodeBodyCfg, nodeRingCfg, ringPickCfg] }, // body, ring, click band
  { heading: "STATE",  cfgs: [selectionRingCfg, hoverRingCfg, reachSphereCfg] },
  { heading: "GUIDES", cfgs: [ringsCfg, handholdsCfg] },
  { heading: "POLES",  cfgs: [scenePolesCfg, nodePolesCfg, selSpherePolesCfg] },
  { heading: "LABELS", cfgs: [globalLabelsCfg] },
];
