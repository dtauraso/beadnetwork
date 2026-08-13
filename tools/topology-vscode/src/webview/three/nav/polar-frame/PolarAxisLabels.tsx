import { AxisLabel } from "./axis-label";
import { THETA_CIRCLES, THETA_CIRCLES_PHI_HALF, PHI_CIRCLES } from "./polar-frame-data";

export function PolarAxisLabels({ poleLen, coneH, arcMid, sfx, octants }: {
  poleLen: number; coneH: number; arcMid: number; sfx: string; octants?: boolean;
}) {
  return (
    <>
      {}
      <AxisLabel text={`+Y pole${sfx}`} color="#22dd55" position={[0, poleLen + coneH * 2, 0]} size={poleLen * 0.12} />
      <AxisLabel text={`+X φ0${sfx}`} color="#dd3333" position={[poleLen + coneH * 2, 0, 0]} size={poleLen * 0.12} />
      <AxisLabel text={`+Z φπ/2${sfx}`} color="#3366dd" position={[0, 0, poleLen + coneH * 2]} size={poleLen * 0.12} />
      {octants && (<>
        <AxisLabel text={`−Y${sfx}`} color="#22dd55" position={[0, -(poleLen + coneH * 2), 0]} size={poleLen * 0.12} />
        <AxisLabel text={`−X φπ${sfx}`} color="#dd3333" position={[-(poleLen + coneH * 2), 0, 0]} size={poleLen * 0.12} />
        <AxisLabel text={`−Z φ3π/2${sfx}`} color="#3366dd" position={[0, 0, -(poleLen + coneH * 2)]} size={poleLen * 0.12} />
      </>)}
      {!octants && (<>
      <AxisLabel text="θ" color="#dd33cc" position={[arcMid, arcMid, 0]} size={poleLen * 0.14} />
      <AxisLabel text="φ" color="#dddd22" position={[arcMid, 0, arcMid]} size={poleLen * 0.14} />
      </>)}
      {octants && THETA_CIRCLES.map((t) => (
        <AxisLabel key={`tl-${t.n}`} text={`${t.sy > 0 ? "+" : "−"}θ`} color={t.c} position={[t.sx * arcMid, t.sy * arcMid, 0]} size={poleLen * 0.11} />
      ))}
      {octants && PHI_CIRCLES.map((p) => (
        <AxisLabel key={`pl-${p.n}`} text={`${p.sz > 0 ? "+" : "−"}φ`} color={p.c} position={[p.sx * arcMid, 0, p.sz * arcMid]} size={poleLen * 0.11} />
      ))}
      {/* Theta again, on the other meridian — named with its phi so the two
          theta circles are told apart. */}
      {octants && THETA_CIRCLES_PHI_HALF.map((m) => (
        <AxisLabel key={`ml-${m.n}`} text={`${m.sy > 0 ? "+" : "−"}θ φπ/2`} color={m.c} position={[0, m.sy * arcMid, m.sz * arcMid]} size={poleLen * 0.11} />
      ))}
    </>
  );
}
