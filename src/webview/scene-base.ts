import { onSpawnRestart } from "./spawn-gen";

declare global {
  interface Window {
    WIREFOLD_ANCHOR_BASE?: string;
    WIREFOLD_SCENE_BASES?: Record<string, string>;
  }
}

let seq = 0;

async function readSelectedBase(): Promise<void> {
  const anchor = window.WIREFOLD_ANCHOR_BASE;
  const bases = window.WIREFOLD_SCENE_BASES;
  if (!anchor || !bases) return;
  try {
    const res = await fetch(`${anchor}/view/scene/selected.bin?r=${++seq}`, { cache: "no-store" });
    if (!res.ok) return;
    const next = bases[(await res.text()).trim()];
    if (next && window.WIREFOLD_SCENE_BASE !== next) window.WIREFOLD_SCENE_BASE = next;
  } catch { /* unreadable; keep the base we have */ }
}

export function startSceneBaseReads(): void {
  if (typeof window === "undefined") return;

  onSpawnRestart(() => void readSelectedBase());
}
