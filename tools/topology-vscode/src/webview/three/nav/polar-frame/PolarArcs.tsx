import { THETA_CIRCLES, PHI_CIRCLES } from "./polar-frame-data";

export function PolarArcs({ arcR, arcTube, octants }: {
  arcR: number; arcTube: number; octants?: boolean;
}) {
  return (
    <>
      {!octants && (<>
      {}
      <mesh raycast={() => null}>
        <torusGeometry args={[arcR, arcTube, 8, 48, Math.PI / 2]} />
        <meshBasicMaterial color="#dd33cc" depthWrite={false} />
      </mesh>
      {}
      <mesh rotation={[Math.PI / 2, 0, 0]} raycast={() => null}>
        <torusGeometry args={[arcR, arcTube, 8, 48, Math.PI / 2]} />
        <meshBasicMaterial color="#dddd22" depthWrite={false} />
      </mesh>
      </>)}
      {octants && THETA_CIRCLES.map((t) => (
        <group key={`tc-${t.n}`} scale={[t.sx, t.sy, 1]}>
          <mesh raycast={() => null}>
            <torusGeometry args={[arcR, arcTube, 8, 48, Math.PI / 2]} />
            <meshBasicMaterial color={t.c} depthWrite={false} />
          </mesh>
        </group>
      ))}
      {octants && PHI_CIRCLES.map((p) => (
        <group key={`pc-${p.n}`} scale={[p.sx, 1, p.sz]}>
          <mesh rotation={[Math.PI / 2, 0, 0]} raycast={() => null}>
            <torusGeometry args={[arcR, arcTube, 8, 48, Math.PI / 2]} />
            <meshBasicMaterial color={p.c} depthWrite={false} />
          </mesh>
        </group>
      ))}
    </>
  );
}
