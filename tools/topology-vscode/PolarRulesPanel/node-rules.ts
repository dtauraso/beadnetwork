import { useSyncExternalStore } from "react";
import { getNodeFrame, subscribeNodeStreamBlocks } from "../src/webview/three/scene/nodes/node-frame-aggregate";
import { getEdgeStreamAccessor } from "../src/webview/three/scene/edges/edge-stream-blocks";
import { nodeKindName } from "../Node/node-kind";
import {
  readNodeDragRLocked,
  readNodeDragPhiLocked,
  readNodeDragThetaMax,
  readNodeDragActive,
  readNodeSelfRLocked,
  readNodeSelfPhiLocked,
  readNodeSelfThetaMax,
  readNodeSelfActive,
  readNodeHasKindRule,
  readNodeKindRuleActive,
  readNodeRuleGroupId,
  readNodeRuleGroupSize,
} from "../Buffer/buffer-layout";
import { nodeLabel } from "../src/webview/three/decode/buffer-decode-node";
import { ruleRowsEqual } from "./node-rules-equal";

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

  edgeRow: number;

  active: boolean;
}

export interface NodeRuleRow {
  row: number;
  label: string;

  kind: string;

  hasRule: boolean;

  rLocked: boolean;

  hasKindRule: boolean;

  kindActive: boolean;

  active: boolean;

  phiLocked: boolean;

  maxThetaPi: number | null;

  hasSelfRule: boolean;

  selfRLocked: boolean;

  selfActive: boolean;

  selfPhiLocked: boolean;

  selfMaxThetaPi: number | null;

  holders: RuleHolder[];

  partners: EdgePartner[];

  groupId: number;

  groupSize: number;
}

const RAD_TO_PI = 1 / Math.PI;

function kindNameFor(nodeView: DataView, row: number): string {
  return nodeKindName(nodeView, row);
}

let cachedRuleRows: NodeRuleRow[] | null = null;


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
    const active = edges.dragActive(edgeRow);
    add(src, { otherRow: dst, otherLabel: nodeLabel(decoded, dst), incoming: false, r, edgeRow, active });
    add(dst, { otherRow: src, otherLabel: nodeLabel(decoded, src), incoming: true, r, edgeRow, active });
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
    const thetaMax = readNodeDragThetaMax(nodeView, row);
    const selfThetaMax = readNodeSelfThetaMax(nodeView, row);

    const rLocked = !!readNodeDragRLocked(nodeView, row);
    const phiLocked = !!readNodeDragPhiLocked(nodeView, row);
    const selfRLocked = !!readNodeSelfRLocked(nodeView, row);
    const selfPhiLocked = !!readNodeSelfPhiLocked(nodeView, row);

    next.push({
      row,
      label: nodeLabel(decoded, row),
      kind: kindNameFor(nodeView, row),
      hasRule: rLocked || phiLocked || thetaMax >= 0,
      rLocked,
      hasKindRule: !!readNodeHasKindRule(nodeView, row),
      kindActive: !!readNodeKindRuleActive(nodeView, row),
      active: !!readNodeDragActive(nodeView, row),
      phiLocked,
      maxThetaPi: thetaMax < 0 ? null : thetaMax * RAD_TO_PI,
      hasSelfRule: selfRLocked || selfPhiLocked || selfThetaMax >= 0,
      selfRLocked,
      selfActive: !!readNodeSelfActive(nodeView, row),
      selfPhiLocked,
      selfMaxThetaPi: selfThetaMax < 0 ? null : selfThetaMax * RAD_TO_PI,
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
