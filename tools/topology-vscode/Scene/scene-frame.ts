import {
  readSceneCX, readSceneCY, readSceneCZ, readSceneRadius,
  readSceneConstantR, readSceneMaxIndexPhi, readSceneMaxIndexTheta,
} from "../Buffer/buffer-layout";

export interface SceneSteps {
  centerX: number;
  centerY: number;
  centerZ: number;

  constantR: number;

  constantPhi: number;
  constantTheta: number;
}

export function sceneSteps(sceneView: DataView): SceneSteps {
  const maxPhi = readSceneMaxIndexPhi(sceneView);
  const maxTheta = readSceneMaxIndexTheta(sceneView);
  return {
    centerX: readSceneCX(sceneView),
    centerY: readSceneCY(sceneView),
    centerZ: readSceneCZ(sceneView),
    constantR: readSceneConstantR(sceneView),
    constantPhi: maxPhi === 0 ? 0 : (2 * Math.PI) / maxPhi,
    constantTheta: maxTheta === 0 ? 0 : (2 * Math.PI) / maxTheta,
  };
}

export function sceneRadius(sceneView: DataView): number {
  return readSceneRadius(sceneView);
}
