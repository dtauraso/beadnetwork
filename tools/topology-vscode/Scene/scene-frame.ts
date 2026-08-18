import {
  readSceneCX, readSceneCY, readSceneCZ, readSceneRadius,
  readSceneConstantR, readSceneMaxIndexPhi, readSceneMaxIndexTheta,
} from "../Buffer/buffer-layout";
import { columnF32, columnI32, hasColumn } from "../Buffer/column-values";
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

export function sceneSteps(sceneView: DataView): SceneSteps {
  const onColumns = hasColumn(COL_STREAM_SCENE_CX);

  const maxPhi = onColumns
    ? columnI32(COL_STREAM_SCENE_MAX_INDEX_PHI)
    : readSceneMaxIndexPhi(sceneView);
  const maxTheta = onColumns
    ? columnI32(COL_STREAM_SCENE_MAX_INDEX_THETA)
    : readSceneMaxIndexTheta(sceneView);

  return {
    centerX: onColumns ? columnF32(COL_STREAM_SCENE_CX) : readSceneCX(sceneView),
    centerY: onColumns ? columnF32(COL_STREAM_SCENE_CY) : readSceneCY(sceneView),
    centerZ: onColumns ? columnF32(COL_STREAM_SCENE_CZ) : readSceneCZ(sceneView),
    constantR: onColumns ? columnF32(COL_STREAM_SCENE_CONSTANT_R) : readSceneConstantR(sceneView),
    constantPhi: maxPhi === 0 ? 0 : (2 * Math.PI) / maxPhi,
    constantTheta: maxTheta === 0 ? 0 : (2 * Math.PI) / maxTheta,
  };
}

export function sceneRadius(sceneView: DataView): number {
  return hasColumn(COL_STREAM_SCENE_RADIUS)
    ? columnF32(COL_STREAM_SCENE_RADIUS)
    : readSceneRadius(sceneView);
}
