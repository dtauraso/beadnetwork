import { columnI32 } from "./column-values";
import {
  COLUMNS_IN_SINGLETON_STREAMS, COLUMNS_PER_NODE_STREAM, COLUMNS_PER_EDGE_STREAM,
} from "./column-streams-gen";
import {
  COL_STREAM_SCENE_NODE_COUNT, COL_STREAM_SCENE_EDGE_COUNT,
} from "../../Scene/columns-gen";

export function nodeColumn(row: number, col: number): number {
  return COLUMNS_IN_SINGLETON_STREAMS + row * COLUMNS_PER_NODE_STREAM + col;
}

export function edgeColumn(row: number, col: number): number {
  const nodes = columnI32(COL_STREAM_SCENE_NODE_COUNT);
  return COLUMNS_IN_SINGLETON_STREAMS
    + nodes * COLUMNS_PER_NODE_STREAM
    + row * COLUMNS_PER_EDGE_STREAM
    + col;
}

export function ownerCounts(): { nodes: number; edges: number } {
  return {
    nodes: columnI32(COL_STREAM_SCENE_NODE_COUNT),
    edges: columnI32(COL_STREAM_SCENE_EDGE_COUNT),
  };
}
