import { useSyncExternalStore } from "react";
import { getNodeFrame, subscribeNodeStreamBlocks } from "../../scene/nodes/node-frame-aggregate";
import { getEdgeStreamAccessor } from "../../scene/edges/edge-stream-blocks";
import { NODE_KIND_NAMES } from "../../../../schema/node-defs";
import { UNKNOWN_KIND_ID } from "../../../../schema/buffer-layout/buffer-layout";
import {
  readNodeKindId,
  readNodeDragRLocked,
  readNodeDragPhiLocked,
  readNodeDragThetaMax,
  readNodeDragActive,
  readNodeHasKindRule,
  readNodeRuleGroupId,
  readNodeRuleGroupSize,
} from "../../../../schema/buffer-layout/buffer-layout";
import { nodeLabel } from "../../decode/buffer-decode-node";

export interface RuleHolder {
  holderRow: number;
  holderLabel: string;

  r: number;
}

export interface EdgePartner {
  otherRow: number;
  otherLabel: string;

  incoming: boolean;

  r: number;
}

export interface NodeRuleRow {
  row: number;
  label: string;

  kind: string;

  hasRule: boolean;

  hasKindRule: boolean;

  active: boolean;

  phiLocked: boolean;

  maxThetaDeg: number | null;

  holders: RuleHolder[];

  partners: EdgePartner[];

  groupId: number;

  groupSize: number;
}

const RAD_TO_DEG = 180 / Math.PI;

function kindNameFor(nodeView: DataView, row: number): string {
  const kindId = readNodeKindId(nodeView, row);
  if (kindId === UNKNOWN_KIND_ID) return "";
  return NODE_KIND_NAMES[kindId] ?? "";
}

let cachedRuleRows: NodeRuleRow[] | null = null;

function holdersEqual(a: RuleHolder[], b: RuleHolder[]): boolean {
  if (a.length !== b.length) return false;
  for (let i = 0; i < a.length; i++) {
    const ai = a[i];
    const bi = b[i];
    if (!ai || !bi) return false;
    if (ai.holderRow !== bi.holderRow || ai.holderLabel !== bi.holderLabel || ai.r !== bi.r) {
      return false;
    }
  }
  return true;
}

function partnersEqual(a: EdgePartner[], b: EdgePartner[]): boolean {
  if (a.length !== b.length) return false;
  for (let i = 0; i < a.length; i++) {
    const ai = a[i];
    const bi = b[i];
    if (!ai || !bi) return false;
    if (
      ai.otherRow !== bi.otherRow ||
      ai.otherLabel !== bi.otherLabel ||
      ai.incoming !== bi.incoming ||
      ai.r !== bi.r
    ) {
      return false;
    }
  }
  return true;
}

function ruleRowsEqual(a: NodeRuleRow[], b: NodeRuleRow[]): boolean {
  if (a.length !== b.length) return false;
  for (let i = 0; i < a.length; i++) {
    const ai = a[i];
    const bi = b[i];
    if (!ai || !bi) return false;
    if (
      ai.row !== bi.row ||
      ai.label !== bi.label ||
      ai.kind !== bi.kind ||
      ai.hasRule !== bi.hasRule ||
      ai.hasKindRule !== bi.hasKindRule ||
      ai.active !== bi.active ||
      ai.phiLocked !== bi.phiLocked ||
      ai.maxThetaDeg !== bi.maxThetaDeg ||
      ai.groupId !== bi.groupId ||
      ai.groupSize !== bi.groupSize ||
      !holdersEqual(ai.holders, bi.holders) ||
      !partnersEqual(ai.partners, bi.partners)
    ) {
      return false;
    }
  }
  return true;
}

function holdersByNode(decoded: ReturnType<typeof getNodeFrame>): Map<number, RuleHolder[]> {
  const byNode = new Map<number, RuleHolder[]>();
  const edges = getEdgeStreamAccessor();
  if (!edges || !decoded) return byNode;
  for (let edgeRow = 0; edgeRow < edges.edgeCount; edgeRow++) {
    const src = edges.srcNodeRow(edgeRow);
    const dst = edges.dstNodeRow(edgeRow);
    if (src < 0 || dst < 0) continue;
    const holder: RuleHolder = {
      holderRow: src,
      holderLabel: nodeLabel(decoded, src),
      r: Math.abs(edges.deltaR(edgeRow)),
    };
    const holders = byNode.get(dst);
    if (holders) {
      if (!holders.some((h) => h.holderRow === src)) holders.push(holder);
    } else {
      byNode.set(dst, [holder]);
    }
  }
  return byNode;
}

function partnersByNode(decoded: ReturnType<typeof getNodeFrame>): Map<number, EdgePartner[]> {
  const byNode = new Map<number, EdgePartner[]>();
  const edges = getEdgeStreamAccessor();
  if (!edges || !decoded) return byNode;

  const add = (row: number, p: EdgePartner) => {
    const list = byNode.get(row);
    if (!list) {
      byNode.set(row, [p]);
      return;
    }
    if (!list.some((q) => q.otherRow === p.otherRow && q.incoming === p.incoming)) list.push(p);
  };

  for (let edgeRow = 0; edgeRow < edges.edgeCount; edgeRow++) {
    const src = edges.srcNodeRow(edgeRow);
    const dst = edges.dstNodeRow(edgeRow);
    if (src < 0 || dst < 0) continue;
    const r = Math.abs(edges.deltaR(edgeRow));
    add(src, { otherRow: dst, otherLabel: nodeLabel(decoded, dst), incoming: false, r });
    add(dst, { otherRow: src, otherLabel: nodeLabel(decoded, src), incoming: true, r });
  }
  return byNode;
}

export function readNodeRuleRows(): NodeRuleRow[] | null {
  const decoded = getNodeFrame();
  if (!decoded) return cachedRuleRows;
  const { nodeCount, nodeView } = decoded;
  const byNode = holdersByNode(decoded);
  const partnerByNode = partnersByNode(decoded);

  const next: NodeRuleRow[] = [];
  for (let row = 0; row < nodeCount; row++) {
    const hasRule = !!readNodeDragRLocked(nodeView, row);
    const thetaMax = readNodeDragThetaMax(nodeView, row);

    next.push({
      row,
      label: nodeLabel(decoded, row),
      kind: kindNameFor(nodeView, row),
      hasRule,
      hasKindRule: !!readNodeHasKindRule(nodeView, row),
      active: !!readNodeDragActive(nodeView, row),
      phiLocked: !!readNodeDragPhiLocked(nodeView, row),
      maxThetaDeg: thetaMax < 0 ? null : thetaMax * RAD_TO_DEG,
      holders: byNode.get(row) ?? [],
      partners: partnerByNode.get(row) ?? [],
      groupId: readNodeRuleGroupId(nodeView, row),
      groupSize: readNodeRuleGroupSize(nodeView, row),
    });
  }

  if (cachedRuleRows && ruleRowsEqual(cachedRuleRows, next)) return cachedRuleRows;
  cachedRuleRows = next;
  return cachedRuleRows;
}

export function useNodeRuleRows(): NodeRuleRow[] | null {
  return useSyncExternalStore(subscribeNodeStreamBlocks, readNodeRuleRows, readNodeRuleRows);
}
