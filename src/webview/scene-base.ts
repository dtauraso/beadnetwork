declare global {
  interface Window {
    WIREFOLD_ANCHOR_BASE?: string;
    WIREFOLD_SCENE_BASES?: Record<string, string>;
  }
}

const READ_INTERVAL_MS = 100;

let seq = 0;
let started = false;

export function startSceneBaseReads(): void {
  if (started || typeof window === "undefined") return;
  started = true;

  const pump = async () => {
    for (;;) {
      const anchor = window.WIREFOLD_ANCHOR_BASE;
      const bases = window.WIREFOLD_SCENE_BASES;
      if (anchor && bases) {
        try {
          const res = await fetch(`${anchor}/view/scene/selected.bin?r=${++seq}`, { cache: "no-store" });
          if (res.ok) {
            const next = bases[(await res.text()).trim()];
            if (next && window.WIREFOLD_SCENE_BASE !== next) window.WIREFOLD_SCENE_BASE = next;
          }
        } catch { /* unreadable this tick; keep the base we have */ }
      }
      await new Promise((r) => setTimeout(r, READ_INTERVAL_MS));
    }
  };
  void pump();
}
