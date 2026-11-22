declare global {
  interface Window {
    BEADNETWORK_ANCHOR_BASE?: string;
    BEADNETWORK_SCENE_BASES?: Record<string, string>;
  }
}

const READ_INTERVAL_MS = 100;


async function readSelectedBase(): Promise<void> {
  const anchor = window.BEADNETWORK_ANCHOR_BASE;
  const bases = window.BEADNETWORK_SCENE_BASES;
  if (!anchor || !bases) return;
  try {
    const res = await fetch(`${anchor}/view/scene/selected.bin`, { cache: "no-store" });
    if (!res.ok) return;
    const next = bases[(await res.text()).trim()];
    if (next && window.BEADNETWORK_SCENE_BASE !== next) window.BEADNETWORK_SCENE_BASE = next;
  } catch { /* unreadable; keep the base we have */ }
}

export function startSceneBaseReads(): void {
  if (typeof window === "undefined") return;

  const pump = async () => {
    for (;;) {
      await readSelectedBase();
      await new Promise((r) => setTimeout(r, READ_INTERVAL_MS));
    }
  };
  void pump();
}
