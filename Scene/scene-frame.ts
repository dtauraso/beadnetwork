import { sceneValue, startSceneReads } from "./scene-leaves";

export interface SceneSteps {
  centerX: number;
  centerY: number;
  centerZ: number;

  constantR: number;

  constantPhi: number;
  constantTheta: number;
}

function read(name: string): number {
  startSceneReads();
  return sceneValue(name);
}

export function sceneSteps(): SceneSteps {
  const maxPhi = read("maxIndexPhi");
  const maxTheta = read("maxIndexTheta");
  return {
    centerX: read("cx"),
    centerY: read("cy"),
    centerZ: read("cz"),
    constantR: read("constantR"),
    constantPhi: maxPhi === 0 ? 0 : (2 * Math.PI) / maxPhi,
    constantTheta: maxTheta === 0 ? 0 : (2 * Math.PI) / maxTheta,
  };
}

export function sceneRadius(): number {
  return read("radius");
}
