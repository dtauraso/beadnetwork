import { PHI_CIRCLES, PHI_CIRCLES_THETA_HALF, THETA_CIRCLES } from "./polar-frame-data";

export function PolarArcs({ arcR, arcTube, octants }: {
  arcR: number; arcTube: number; octants?: boolean;
}) {
  return (
    <>
      {!octants && (<>
      {}
      <mesh raycast={() => null}>
        <torusGeometry args={[arcR, arcTube, 8, 48, Math.PI / 2]} />
        <meshBasicMaterial color="#dd33cc" />
      </mesh>
      {}
      <mesh rotation={[Math.PI / 2, 0, 0]} raycast={() => null}>
        <torusGeometry args={[arcR, arcTube, 8, 48, Math.PI / 2]} />
        <meshBasicMaterial color="#dddd22" />
      </mesh>
      </>)}
      {octants && PHI_CIRCLES.map((t) => (
        <group key={`tc-${t.n}`} scale={[t.sx, t.sy, 1]}>
          <mesh raycast={() => null}>
            <torusGeometry args={[arcR, arcTube, 8, 48, Math.PI / 2]} />
            <meshBasicMaterial color={t.c} />
          </mesh>
        </group>
      ))}
      {/* Rotating -pi/2 about y carries the base arc's +x leg onto +z, so the
          unscaled quarter sits in (+y,+z) and (sy,sz) selects the quadrant the
          same way it does for the other two circles. */}
      {octants && PHI_CIRCLES_THETA_HALF.map((m) => (
        <group key={`mc-${m.n}`} scale={[1, m.sy, m.sz]}>
          <mesh rotation={[0, -Math.PI / 2, 0]} raycast={() => null}>
            <torusGeometry args={[arcR, arcTube, 8, 48, Math.PI / 2]} />
            <meshBasicMaterial color={m.c} />
          </mesh>
        </group>
      ))}
      {octants && THETA_CIRCLES.map((p) => (
        <group key={`pc-${p.n}`} scale={[p.sx, 1, p.sz]}>
          <mesh rotation={[Math.PI / 2, 0, 0]} raycast={() => null}>
            <torusGeometry args={[arcR, arcTube, 8, 48, Math.PI / 2]} />
            <meshBasicMaterial color={p.c} />
          </mesh>
        </group>
      ))}
    </>
  );
}
