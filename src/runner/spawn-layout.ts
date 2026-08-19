import {
  EDGE_BASE_FD,
  MAX_EDGE_STREAMS,
  MAX_NODE_STREAMS,
  MAX_COLUMN_STREAMS,
  VIEW_FD,
} from "./stream-fds";
import {
  columnStreamCount,
} from "../Buffer/column-streams-gen";

export interface SpawnLayout {
  edgeCount: number;
  nodeCount: number;
  nodeBaseFd: number;
  interiorBaseFd: number;

  beadBaseFd: number;

  colBaseFd: number;
  colCount: number;

  stdio: Array<"pipe">;

  streamFDsEnv: string;

  warnings: string[];
}

export function computeSpawnLayout(counts: { nodes: number; edges: number }): SpawnLayout {
  const warnings: string[] = [];

  const edgeCountRaw = counts.edges;
  const edgeCount = edgeCountRaw > MAX_EDGE_STREAMS ? 0 : edgeCountRaw;
  if (edgeCountRaw > MAX_EDGE_STREAMS) {

    warnings.push(
      `edge count ${edgeCountRaw} exceeds MAX_EDGE_STREAMS (${MAX_EDGE_STREAMS}); disabling ALL dedicated per-edge streams for this run`,
    );
  }

  const nodeCountRaw = counts.nodes;
  const nodeCount = nodeCountRaw > MAX_NODE_STREAMS ? 0 : nodeCountRaw;
  if (nodeCountRaw > MAX_NODE_STREAMS) {

    warnings.push(
      `node count ${nodeCountRaw} exceeds MAX_NODE_STREAMS (${MAX_NODE_STREAMS}); disabling ALL dedicated per-node NODE/INTERIOR streams for this run`,
    );
  }
  const nodeBaseFd = EDGE_BASE_FD + edgeCount;
  const interiorBaseFd = nodeBaseFd + nodeCount;

  const beadBaseFd = interiorBaseFd + nodeCount;

  const stdio: Array<"pipe"> = ["pipe", "pipe", "pipe", "pipe", "pipe"];
  for (let i = 0; i < edgeCount; i++) stdio.push("pipe");
  for (let i = 0; i < nodeCount; i++) stdio.push("pipe");
  for (let i = 0; i < nodeCount; i++) stdio.push("pipe");

  for (let i = 0; i < nodeCount; i++) stdio.push("pipe");

  const colBaseFd = beadBaseFd + nodeCount;
  const colCountRaw = columnStreamCount(nodeCount, edgeCount);
  const colCount = colCountRaw > MAX_COLUMN_STREAMS ? 0 : colCountRaw;
  if (colCountRaw > MAX_COLUMN_STREAMS) {

    warnings.push(
      `column-stream count ${colCountRaw} for ${nodeCount} nodes / ${edgeCount} edges exceeds ` +
      `MAX_COLUMN_STREAMS (${MAX_COLUMN_STREAMS}); per-column streams are OFF for this run. ` +
      `Each pipe costs two descriptors and spawn fails outright past roughly 5000 pipes.`,
    );
  }
  for (let i = 0; i < colCount; i++) stdio.push("pipe");

  const streamFDsEnvParts = [`view:${VIEW_FD}`];
  if (edgeCount > 0) streamFDsEnvParts.push(`edge:${EDGE_BASE_FD}`);

  if (nodeCount > 0) {
    streamFDsEnvParts.push(
      `node:${nodeBaseFd}`, `interior:${interiorBaseFd}`, `bead:${beadBaseFd}`,
    );
  }
  if (colCount > 0) streamFDsEnvParts.push(`col:${colBaseFd}`, `colcount:${colCount}`);
  const streamFDsEnv = streamFDsEnvParts.join(",");

  return {
    edgeCount, nodeCount, nodeBaseFd, interiorBaseFd, beadBaseFd,
    colBaseFd, colCount, stdio, streamFDsEnv, warnings,
  };
}
