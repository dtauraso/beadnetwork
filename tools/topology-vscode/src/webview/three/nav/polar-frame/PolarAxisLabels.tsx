import { AxisLabel } from "./axis-label";
import { PHI_CIRCLES, PHI_CIRCLES_THETA_HALF, THETA_CIRCLES } from "./polar-frame-data";

export function PolarAxisLabels({ poleLen, coneH, arcMid, sfx, octants }: {
  poleLen: number; coneH: number; arcMid: number; sfx: string; octants?: boolean;
}) {
  return (
    <>
      {}
      <AxisLabel text={`+Y pole${sfx}`} color="#22dd55" position={[0, poleLen + coneH * 2, 0]} size={poleLen * 0.12} />
      <AxisLabel text={`+X φπ/2 θ0${sfx}`} color="#dd3333" position={[poleLen + coneH * 2, 0, 0]} size={poleLen * 0.12} />
      <AxisLabel text={`+Z φπ/2 θπ/2${sfx}`} color="#3366dd" position={[0, 0, poleLen + coneH * 2]} size={poleLen * 0.12} />
      {octants && (<>
        <AxisLabel text={`−Y φπ${sfx}`} color="#22dd55" position={[0, -(poleLen + coneH * 2), 0]} size={poleLen * 0.12} />
        <AxisLabel text={`−X φπ/2 θπ${sfx}`} color="#dd3333" position={[-(poleLen + coneH * 2), 0, 0]} size={poleLen * 0.12} />
        <AxisLabel text={`−Z φπ/2 θ3π/2${sfx}`} color="#3366dd" position={[0, 0, -(poleLen + coneH * 2)]} size={poleLen * 0.12} />
      </>)}
      {!octants && (<>
      <AxisLabel text="φ" color="#dd33cc" position={[arcMid, arcMid, 0]} size={poleLen * 0.14} />
      <AxisLabel text="θ" color="#dddd22" position={[arcMid, 0, arcMid]} size={poleLen * 0.14} />
      </>)}
      {/* Each arc is named by the interval it covers, in multiples of pi. The
          name comes from the arc's own entry rather than being rebuilt from its
          signs here, so the name and the arc cannot drift apart. */}
      {octants && PHI_CIRCLES.map((t) => (
        <AxisLabel key={`tl-${t.n}`} text={t.label} color={t.c} position={[t.sx * arcMid, t.sy * arcMid, 0]} size={poleLen * 0.11} />
      ))}
      {octants && THETA_CIRCLES.map((p) => (
        <AxisLabel key={`pl-${p.n}`} text={p.label} color={p.c} position={[p.sx * arcMid, 0, p.sz * arcMid]} size={poleLen * 0.11} />
      ))}
      {octants && PHI_CIRCLES_THETA_HALF.map((m) => (
        <AxisLabel key={`ml-${m.n}`} text={m.label} color={m.c} position={[0, m.sy * arcMid, m.sz * arcMid]} size={poleLen * 0.11} />
      ))}
    </>
  );
}
