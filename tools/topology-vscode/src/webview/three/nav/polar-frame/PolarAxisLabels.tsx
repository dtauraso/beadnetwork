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
      {/* Each arc is named by the interval it covers, in multiples of pi. The
          name comes from the arc's own entry rather than being rebuilt from its
          signs here, so the name and the arc cannot drift apart. */}
      {octants && THETA_CIRCLES.map((t) => (
        <AxisLabel key={`tl-${t.n}`} text={t.label} color={t.c} position={[t.sx * arcMid, t.sy * arcMid, 0]} size={poleLen * 0.11} />
      ))}
      {octants && PHI_CIRCLES.map((p) => (
        <AxisLabel key={`pl-${p.n}`} text={p.label} color={p.c} position={[p.sx * arcMid, 0, p.sz * arcMid]} size={poleLen * 0.11} />
      ))}
      {octants && THETA_CIRCLES_PHI_HALF.map((m) => (
        <AxisLabel key={`ml-${m.n}`} text={m.label} color={m.c} position={[0, m.sy * arcMid, m.sz * arcMid]} size={poleLen * 0.11} />
      ))}
    </>
  );
}
