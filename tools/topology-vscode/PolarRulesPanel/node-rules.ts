import { useSyncExternalStore } from "react";
import { getNodeSections, subscribeNodeStreamBlocks } from "../src/webview/three/scene/nodes/node-sections";
import { getEdgeStreamAccessor } from "../src/webview/three/scene/edges/edge-stream-blocks";
import { nodeKindName } from "../Node/node-kind";
import { ownerCounts } from "../Buffer/column-owners";
import { columnF32, columnU8, columnI32 } from "../Buffer/column-values";
import { nodeColumn } from "../Buffer/column-owners";
import {
  COL_STREAM_NODE_DRAG_RLOCKED, COL_STREAM_NODE_DRAG_PHI_LOCKED,
  COL_STREAM_NODE_DRAG_THETA_MAX, COL_STREAM_NODE_DRAG_ACTIVE,
  COL_STREAM_NODE_HAS_KIND_RULE, COL_STREAM_NODE_KIND_RULE_ACTIVE,
  COL_STREAM_NODE_SELF_RLOCKED, COL_STREAM_NODE_SELF_PHI_LOCKED,
  COL_STREAM_NODE_SELF_THETA_MAX, COL_STREAM_NODE_SELF_ACTIVE,
  COL_STREAM_NODE_RULE_GROUP_ID, COL_STREAM_NODE_RULE_GROUP_SIZE,
} from "../Buffer/column-streams-gen";
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

function kindNameFor(row: number): string {
  return nodeKindName(row);
}

let cachedRuleRows: NodeRuleRow[] | null = null;


function holdersByNode(nodeCount: number): Map<number, RuleHolder[]> {
  const byNode = new Map<number, RuleHolder[]>();
  const edges = getEdgeStreamAccessor();
  if (!edges || nodeCount <= 0) return byNode;
  for (let edgeRow = 0; edgeRow < edges.edgeCount; edgeRow++) {
    const src = edges.srcNodeRow(edgeRow);
    const dst = edges.dstNodeRow(edgeRow);
    if (src < 0 || dst < 0) continue;
    const holder: RuleHolder = {
      holderRow: src,
      holderLabel: nodeLabel(src),
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

function partnersByNode(nodeCount: number): Map<number, EdgePartner[]> {
  const byNode = new Map<number, EdgePartner[]>();
  const edges = getEdgeStreamAccessor();
  if (!edges || nodeCount <= 0) return byNode;

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
    add(src, { otherRow: dst, otherLabel: nodeLabel(dst), incoming: false, r, edgeRow, active });
    add(dst, { otherRow: src, otherLabel: nodeLabel(src), incoming: true, r, edgeRow, active });
  }
  return byNode;
}

export function readNodeRuleRows(): NodeRuleRow[] | null {
  const nodeCount = ownerCounts().nodes;
  if (nodeCount <= 0) return cachedRuleRows;
  const byNode = holdersByNode(nodeCount);
  const partnerByNode = partnersByNode(nodeCount);

  const next: NodeRuleRow[] = [];
  for (let row = 0; row < nodeCount; row++) {
    const thetaMax = columnF32(nodeColumn(row, COL_STREAM_NODE_DRAG_THETA_MAX));
    const selfThetaMax = columnF32(nodeColumn(row, COL_STREAM_NODE_SELF_THETA_MAX));

    const rLocked = !!columnU8(nodeColumn(row, COL_STREAM_NODE_DRAG_RLOCKED));
    const phiLocked = !!columnU8(nodeColumn(row, COL_STREAM_NODE_DRAG_PHI_LOCKED));
    const selfRLocked = !!columnU8(nodeColumn(row, COL_STREAM_NODE_SELF_RLOCKED));
    const selfPhiLocked = !!columnU8(nodeColumn(row, COL_STREAM_NODE_SELF_PHI_LOCKED));

    next.push({
      row,
      label: nodeLabel(row),
      kind: kindNameFor(row),
      hasRule: rLocked || phiLocked || thetaMax >= 0,
      rLocked,
      hasKindRule: !!columnU8(nodeColumn(row, COL_STREAM_NODE_HAS_KIND_RULE)),
      kindActive: !!columnU8(nodeColumn(row, COL_STREAM_NODE_KIND_RULE_ACTIVE)),
      active: !!columnU8(nodeColumn(row, COL_STREAM_NODE_DRAG_ACTIVE)),
      phiLocked,
      maxThetaPi: thetaMax < 0 ? null : thetaMax * RAD_TO_PI,
      hasSelfRule: selfRLocked || selfPhiLocked || selfThetaMax >= 0,
      selfRLocked,
      selfActive: !!columnU8(nodeColumn(row, COL_STREAM_NODE_SELF_ACTIVE)),
      selfPhiLocked,
      selfMaxThetaPi: selfThetaMax < 0 ? null : selfThetaMax * RAD_TO_PI,
      holders: byNode.get(row) ?? [],
      partners: partnerByNode.get(row) ?? [],
      groupId: columnI32(nodeColumn(row, COL_STREAM_NODE_RULE_GROUP_ID)),
      groupSize: columnI32(nodeColumn(row, COL_STREAM_NODE_RULE_GROUP_SIZE)),
    });
  }

  if (cachedRuleRows && ruleRowsEqual(cachedRuleRows, next)) return cachedRuleRows;
  cachedRuleRows = next;
  return cachedRuleRows;
}

export function useNodeRuleRows(): NodeRuleRow[] | null {
  return useSyncExternalStore(subscribeNodeStreamBlocks, readNodeRuleRows, readNodeRuleRows);
}
