import { AxisLabel } from "./axis-label";

export function PolarAxisLabels({ poleLen, coneH, arcMid, sfx }: {
  poleLen: number; coneH: number; arcMid: number; sfx: string;
}) {
  return (
    <>
      {}
      <AxisLabel text={`+Y pole${sfx}`} color="#22dd55" position={[0, poleLen + coneH * 2, 0]} size={poleLen * 0.12} />
      <AxisLabel text={`+X θ0${sfx}`} color="#dd3333" position={[poleLen + coneH * 2, 0, 0]} size={poleLen * 0.12} />
      <AxisLabel text={`+Z θπ/2${sfx}`} color="#3366dd" position={[0, 0, poleLen + coneH * 2]} size={poleLen * 0.12} />
      <AxisLabel text="φ" color="#dd33cc" position={[arcMid, arcMid, 0]} size={poleLen * 0.14} />
      <AxisLabel text="θ" color="#dddd22" position={[arcMid, 0, arcMid]} size={poleLen * 0.14} />
    </>
  );
}
