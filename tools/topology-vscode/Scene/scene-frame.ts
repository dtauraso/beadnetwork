import { columnF32, columnI32 } from "../Buffer/column-values";
import {
  COL_STREAM_SCENE_CX, COL_STREAM_SCENE_CY, COL_STREAM_SCENE_CZ, COL_STREAM_SCENE_RADIUS,
  COL_STREAM_SCENE_CONSTANT_R, COL_STREAM_SCENE_MAX_INDEX_PHI, COL_STREAM_SCENE_MAX_INDEX_THETA,
} from "../Buffer/column-streams-gen";

export interface SceneSteps {
  centerX: number;
  centerY: number;
  centerZ: number;

  constantR: number;

  constantPhi: number;
  constantTheta: number;
}

export function sceneSteps(): SceneSteps {
  const maxPhi = columnI32(COL_STREAM_SCENE_MAX_INDEX_PHI);
  const maxTheta = columnI32(COL_STREAM_SCENE_MAX_INDEX_THETA);
  return {
    centerX: columnF32(COL_STREAM_SCENE_CX),
    centerY: columnF32(COL_STREAM_SCENE_CY),
    centerZ: columnF32(COL_STREAM_SCENE_CZ),
    constantR: columnF32(COL_STREAM_SCENE_CONSTANT_R),
    constantPhi: maxPhi === 0 ? 0 : (2 * Math.PI) / maxPhi,
    constantTheta: maxTheta === 0 ? 0 : (2 * Math.PI) / maxTheta,
  };
}

export function sceneRadius(): number {
  return columnF32(COL_STREAM_SCENE_RADIUS);
}
