import React, { useMemo } from "react";
import * as THREE from "three";
import { AxisLabel } from "./axis-label";
import { computePolarFrameGeometry } from "./polar-frame-geometry";
import { HANDHOLD_TERM_TAG, THETA_CIRCLES, PHI_CIRCLES } from "./polar-frame-data";

const WORLD_UP = new THREE.Vector3(0, 1, 0);

export function PolarFrame({ center, scale, tag, octants, pole }: {
  center: THREE.Vector3; scale: number; tag?: string; octants?: boolean; pole?: THREE.Vector3;
}) {
  const { poleLen, poleRadius, coneH, coneBaseR, arcR, arcTube, arcMid, hhR, arcHH } =
    computePolarFrameGeometry(scale);
  const sfx = tag ? ` ${tag}` : "";

  const quat = useMemo(() => {
    const q = new THREE.Quaternion();
    if (pole) q.setFromUnitVectors(WORLD_UP, pole.clone().normalize());
    return q;
  }, [pole]);
  return (
    <group position={[center.x, center.y, center.z]} quaternion={quat}>
      {}
      <mesh position={[0, poleLen / 2, 0]} raycast={() => null}>
        <cylinderGeometry args={[poleRadius, poleRadius, poleLen, 12]} />
        <meshBasicMaterial color="#22dd55" depthWrite={false} />
      </mesh>
      <mesh position={[0, poleLen + coneH / 2, 0]} raycast={() => null}>
        <coneGeometry args={[coneBaseR, coneH, 12]} />
        <meshBasicMaterial color="#22dd55" depthWrite={false} />
      </mesh>
      {}
      <mesh position={[poleLen / 2, 0, 0]} rotation={[0, 0, -Math.PI / 2]} raycast={() => null}>
        <cylinderGeometry args={[poleRadius, poleRadius, poleLen, 12]} />
        <meshBasicMaterial color="#dd3333" depthWrite={false} />
      </mesh>
      <mesh position={[poleLen + coneH / 2, 0, 0]} rotation={[0, 0, -Math.PI / 2]} raycast={() => null}>
        <coneGeometry args={[coneBaseR, coneH, 12]} />
        <meshBasicMaterial color="#dd3333" depthWrite={false} />
      </mesh>
      {}
      <mesh position={[0, 0, poleLen / 2]} rotation={[Math.PI / 2, 0, 0]} raycast={() => null}>
        <cylinderGeometry args={[poleRadius, poleRadius, poleLen, 12]} />
        <meshBasicMaterial color="#3366dd" depthWrite={false} />
      </mesh>
      <mesh position={[0, 0, poleLen + coneH / 2]} rotation={[Math.PI / 2, 0, 0]} raycast={() => null}>
        <coneGeometry args={[coneBaseR, coneH, 12]} />
        <meshBasicMaterial color="#3366dd" depthWrite={false} />
      </mesh>
      {}
      {octants && (<>
        <mesh position={[0, -poleLen / 2, 0]} raycast={() => null}>
          <cylinderGeometry args={[poleRadius, poleRadius, poleLen, 12]} />
          <meshBasicMaterial color="#22dd55" depthWrite={false} />
        </mesh>
        <mesh position={[0, -(poleLen + coneH / 2), 0]} rotation={[Math.PI, 0, 0]} raycast={() => null}>
          <coneGeometry args={[coneBaseR, coneH, 12]} />
          <meshBasicMaterial color="#22dd55" depthWrite={false} />
        </mesh>
        <mesh position={[-poleLen / 2, 0, 0]} rotation={[0, 0, Math.PI / 2]} raycast={() => null}>
          <cylinderGeometry args={[poleRadius, poleRadius, poleLen, 12]} />
          <meshBasicMaterial color="#dd3333" depthWrite={false} />
        </mesh>
        <mesh position={[-(poleLen + coneH / 2), 0, 0]} rotation={[0, 0, Math.PI / 2]} raycast={() => null}>
          <coneGeometry args={[coneBaseR, coneH, 12]} />
          <meshBasicMaterial color="#dd3333" depthWrite={false} />
        </mesh>
        <mesh position={[0, 0, -poleLen / 2]} rotation={[Math.PI / 2, 0, 0]} raycast={() => null}>
          <cylinderGeometry args={[poleRadius, poleRadius, poleLen, 12]} />
          <meshBasicMaterial color="#3366dd" depthWrite={false} />
        </mesh>
        <mesh position={[0, 0, -(poleLen + coneH / 2)]} rotation={[-Math.PI / 2, 0, 0]} raycast={() => null}>
          <coneGeometry args={[coneBaseR, coneH, 12]} />
          <meshBasicMaterial color="#3366dd" depthWrite={false} />
        </mesh>
      </>)}
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
      {octants && (<>
        {}
        {([[arcR, 0, 0], [-arcR, 0, 0], [0, arcR, 0], [0, -arcR, 0], [0, 0, arcR], [0, 0, -arcR]] as [number, number, number][]).map((p, i) => (
          <mesh key={`hhp-${i}`} position={p} userData={{ [HANDHOLD_TERM_TAG]: true }}>
            <sphereGeometry args={[hhR, 12, 12]} />
            <meshStandardMaterial color="#cc8844" emissive="#cc8844" emissiveIntensity={0.6} />
          </mesh>
        ))}
        {}
        {THETA_CIRCLES.map((t) => (
          <mesh
            key={`th-${t.n}`}
            position={[t.sx * arcHH, t.sy * arcHH, 0]}
            userData={{ [HANDHOLD_TERM_TAG]: true }}
          >
            <sphereGeometry args={[hhR, 12, 12]} />
            <meshStandardMaterial color="#cc8844" emissive="#cc8844" emissiveIntensity={0.6} />
          </mesh>
        ))}
        {PHI_CIRCLES.map((p) => (
          <mesh
            key={`ph-${p.n}`}
            position={[p.sx * arcHH, 0, p.sz * arcHH]}
            userData={{ [HANDHOLD_TERM_TAG]: true }}
          >
            <sphereGeometry args={[hhR, 12, 12]} />
            <meshStandardMaterial color="#cc8844" emissive="#cc8844" emissiveIntensity={0.6} />
          </mesh>
        ))}
      </>)}
    </group>
  );
}
