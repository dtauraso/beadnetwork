import {
  EDGE_BASE_FD,
  MAX_EDGE_STREAMS,
  MAX_NODE_STREAMS,
  VIEW_FD,
} from "./stream-fds";

export interface SpawnLayout {
  edgeCount: number;
  nodeCount: number;
  nodeBaseFd: number;
  interiorBaseFd: number;

  beadBaseFd: number;

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

  const streamFDsEnvParts = [`view:${VIEW_FD}`];
  if (edgeCount > 0) streamFDsEnvParts.push(`edge:${EDGE_BASE_FD}`);

  if (nodeCount > 0) {
    streamFDsEnvParts.push(
      `node:${nodeBaseFd}`, `interior:${interiorBaseFd}`, `bead:${beadBaseFd}`,
    );
  }
  const streamFDsEnv = streamFDsEnvParts.join(",");

  return {
    edgeCount, nodeCount, nodeBaseFd, interiorBaseFd, beadBaseFd,
    stdio, streamFDsEnv, warnings,
  };
}
