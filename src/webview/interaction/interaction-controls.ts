import { useCallback } from "react";
import * as THREE from "three";
import { sendRawInput, buildPointerRaw, buildWheelRaw } from "./raw-input";
import type { PickRef } from "./pick-types";
import { postLog } from "../log/post";

export function useInteractionControls(
  cameraRef: React.MutableRefObject<THREE.PerspectiveCamera | null>,
  pickRequest: PickRef,
) {
  const onPointerDown = useCallback((e: React.PointerEvent<HTMLDivElement>) => {
    const ev = buildPointerRaw(e, "pointerdown", cameraRef, pickRequest);
    postLog("ts-pointer-down", {
      handler: true,
      built: ev !== null,
      cam: cameraRef.current !== null,
      pick: pickRequest.current !== null,
      hit: ev?.hit.kind ?? "",
    });
    if (ev) sendRawInput(ev);
    e.currentTarget.setPointerCapture(e.pointerId);
  }, [cameraRef, pickRequest]);

  const onPointerMove = useCallback((e: React.PointerEvent<HTMLDivElement>) => {
    const ev = buildPointerRaw(e, "pointermove", cameraRef, pickRequest);
    if (ev) sendRawInput(ev);
  }, [cameraRef, pickRequest]);

  const onPointerUp = useCallback((e: React.PointerEvent<HTMLDivElement>) => {
    const ev = buildPointerRaw(e, "pointerup", cameraRef, pickRequest);
    if (ev) sendRawInput(ev);
    e.currentTarget.releasePointerCapture(e.pointerId);
  }, [cameraRef, pickRequest]);

  const onPointerCancel = useCallback((e: React.PointerEvent<HTMLDivElement>) => {
    const ev = buildPointerRaw(e, "pointerup", cameraRef, pickRequest);
    if (ev) sendRawInput(ev);
  }, [cameraRef, pickRequest]);

  const onWheelNative = useCallback((e: WheelEvent) => {
    e.preventDefault();
    const ev = buildWheelRaw(e, cameraRef, pickRequest);
    if (ev) sendRawInput(ev);
  }, [cameraRef, pickRequest]);

  return { onPointerDown, onPointerMove, onPointerUp, onPointerCancel, onWheelNative };
}
