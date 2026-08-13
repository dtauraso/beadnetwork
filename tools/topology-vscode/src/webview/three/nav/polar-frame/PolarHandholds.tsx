import { HANDHOLD_TERM_TAG, PHI_CIRCLES, THETA_CIRCLES } from "./polar-frame-data";

export function PolarHandholds({ arcR, arcHH, hhR }: {
  arcR: number; arcHH: number; hhR: number;
}) {
  return (
    <>
      {}
      {([[arcR, 0, 0], [-arcR, 0, 0], [0, arcR, 0], [0, -arcR, 0], [0, 0, arcR], [0, 0, -arcR]] as [number, number, number][]).map((p, i) => (
        <mesh key={`hhp-${i}`} position={p} userData={{ [HANDHOLD_TERM_TAG]: true }}>
          <sphereGeometry args={[hhR, 12, 12]} />
          <meshStandardMaterial color="#cc8844" emissive="#cc8844" emissiveIntensity={0.6} />
        </mesh>
      ))}
      {}
      {PHI_CIRCLES.map((t) => (
        <mesh
          key={`th-${t.n}`}
          position={[t.sx * arcHH, t.sy * arcHH, 0]}
          userData={{ [HANDHOLD_TERM_TAG]: true }}
        >
          <sphereGeometry args={[hhR, 12, 12]} />
          <meshStandardMaterial color="#cc8844" emissive="#cc8844" emissiveIntensity={0.6} />
        </mesh>
      ))}
      {THETA_CIRCLES.map((p) => (
        <mesh
          key={`ph-${p.n}`}
          position={[p.sx * arcHH, 0, p.sz * arcHH]}
          userData={{ [HANDHOLD_TERM_TAG]: true }}
        >
          <sphereGeometry args={[hhR, 12, 12]} />
          <meshStandardMaterial color="#cc8844" emissive="#cc8844" emissiveIntensity={0.6} />
        </mesh>
      ))}
    </>
  );
}
