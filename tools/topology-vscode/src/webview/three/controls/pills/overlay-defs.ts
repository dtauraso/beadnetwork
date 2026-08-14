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

const sceneVectorsCfg: ToggleCfg = {
  flag: "sceneVectors",
  default: true,
  active: (v) => v,
  icon: "↗",
  label: "scene vectors",
  title: (a) => (a ? "Hide scene-centre vectors" : "Show scene-centre vectors"),
  payload: (v) => ({ flag: "sceneVectors", was: v }),
};

const commEdgesCfg: ToggleCfg = {
  flag: "commEdges",
  default: false,
  active: (v) => v,
  icon: "⇢",
  label: "comm edges",
  title: (a) =>
    a ? "Hide constraint-communication edges" : "Show constraint-communication edges",
  payload: (v) => ({ flag: "commEdges", was: v }),
};

export type OverlayGroup = {
  heading: string;
  cfgs: ToggleCfg[];

  groups?: OverlayGroup[];
};

export function groupCfgs(group: OverlayGroup): ToggleCfg[] {
  return [...group.cfgs, ...(group.groups ?? []).flatMap(groupCfgs)];
}

export const OVERLAY_GROUPS: OverlayGroup[] = [
  {
    heading: "NODE",
    cfgs: [],
    groups: [
      { heading: "SHAPE", cfgs: [nodeBodyCfg, nodeRingCfg, ringPickCfg] },
      { heading: "STATE", cfgs: [selectionRingCfg, hoverRingCfg] },
      { heading: "REACH", cfgs: [reachSphereCfg, selSpherePolesCfg] },
      { heading: "POLES", cfgs: [nodePolesCfg] },
    ],
  },
  {
    heading: "SCENE",
    cfgs: [],
    groups: [
      { heading: "GUIDES", cfgs: [ringsCfg, handholdsCfg] },
      { heading: "POLES", cfgs: [scenePolesCfg] },
      { heading: "VECTORS", cfgs: [sceneVectorsCfg, commEdgesCfg] },
      { heading: "LABELS", cfgs: [globalLabelsCfg] },
    ],
  },
];
