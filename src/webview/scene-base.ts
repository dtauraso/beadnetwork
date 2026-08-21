import { SCENES } from "../Scene/scenes-gen";

declare global {
  interface Window {
    WIREFOLD_CONTAINER_BASE?: string;
    WIREFOLD_ANCHOR_BASE?: string;
  }
}

const READ_INTERVAL_MS = 250;

let seq = 0;
let started = false;

export function startSceneBaseReads(): void {
  if (started || typeof window === "undefined") return;
  started = true;

  const pump = async () => {
    for (;;) {
      const anchor = window.WIREFOLD_ANCHOR_BASE;
      const container = window.WIREFOLD_CONTAINER_BASE;
      if (anchor && container) {
        try {
          const res = await fetch(`${anchor}/view/scene/selected.bin?r=${++seq}`, { cache: "no-store" });
          if (res.ok) {
            const selected = (await res.text()).trim();
            const scene = SCENES.find((s) => s.name === selected) ?? SCENES[0];
            if (scene) {
              const next = `${container}/${scene.dir}`;
              if (window.WIREFOLD_SCENE_BASE !== next) window.WIREFOLD_SCENE_BASE = next;
            }
          }
        } catch { /* the selection is unreadable this tick; keep the base we have */ }
      }
      await new Promise((r) => setTimeout(r, READ_INTERVAL_MS));
    }
  };
  void pump();
}
