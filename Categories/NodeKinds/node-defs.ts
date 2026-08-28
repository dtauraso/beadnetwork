
import type { NodeDef } from "./node-def";
import * as NodePhi from "./NodePhi/node-def-gen";
import * as NodePhiTheta from "./NodePhiTheta/node-def-gen";
import * as NodePhiTheta3 from "./NodePhiTheta3/node-def-gen";
import * as PulseLeft from "./PulseLeft/node-def-gen";
import * as PulseRight from "./PulseRight/node-def-gen";
import * as Time from "./Time/node-def-gen";
import * as TimeEnd from "./TimeEnd/node-def-gen";
import * as TimeStart from "./TimeStart/node-def-gen";
import * as Input from "./input/node-def-gen";
import * as Pulse from "./pulse/node-def-gen";
import * as SelectLeft from "./selectleft/node-def-gen";
import * as SelectRight from "./selectright/node-def-gen";

export type { NodeDef };

interface KindFragment {
  KIND_NAME: string;
  KIND_ID: number;
  DEF: NodeDef;
}

const KIND_FRAGMENTS: readonly KindFragment[] = [
  Input,
  NodePhi,
  NodePhiTheta,
  NodePhiTheta3,
  Pulse,
  PulseLeft,
  PulseRight,
  SelectLeft,
  SelectRight,
  Time,
  TimeEnd,
  TimeStart,
];

export const RUNTIME_IMPLEMENTED_KINDS: ReadonlySet<string> = new Set(
  KIND_FRAGMENTS.map((k) => k.KIND_NAME),
);

export const NODE_DEFS: Record<string, NodeDef> = Object.fromEntries(
  KIND_FRAGMENTS.map((k) => [k.KIND_NAME, k.DEF]),
);

/** The KindId a node carries when its kind is not in the registry — the one
 * index NODE_DEFS_ARRAY below must never be asked for. Matches KindIDUnknown
 * in Categories/Node/node_kind_id_gen.go. */
export const UNKNOWN_KIND_ID = 0xff;

const MAX_KIND_ID = KIND_FRAGMENTS.reduce((m, k) => Math.max(m, k.KIND_ID), 0);

function byKindId<T>(pick: (k: KindFragment) => T): readonly (T | undefined)[] {
  const out = new Array<T | undefined>(MAX_KIND_ID + 1).fill(undefined);
  for (const k of KIND_FRAGMENTS) out[k.KIND_ID] = pick(k);
  return out;
}

export const NODE_DEFS_ARRAY: readonly (NodeDef | undefined)[] = byKindId(
  (k) => k.DEF,
);

export const NODE_KIND_NAMES: readonly (string | undefined)[] = byKindId(
  (k) => k.KIND_NAME,
);
