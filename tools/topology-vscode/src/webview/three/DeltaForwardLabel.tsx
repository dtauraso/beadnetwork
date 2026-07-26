import { createPortal } from "react-dom";
import { useDeltaForwardRows } from "./overlay-flags";

// DeltaForwardLabel — the in-editor "delta forwarded" log, right of AbcDragLabel
// (portals into #delta-forward-mount, html.ts, which sits to the right of
// #abc-drag-mount in the toolbar's drag-log-row). Reflects the delta-forward
// full-graph propagation feature: every node relays the delta triple it FIRST
// picks up (from the dragged node itself, or from a neighbor that already
// forwarded) to its OTHER neighbors EXACTLY ONCE per drag
// (moveMsgKindDeltaForward, nodes/Wiring/node_mover.go's forwardDeltaOnce,
// guarded by forwardedThisDrag) — so the triple spreads across the whole
// reachable graph via independent concurrent single hops, terminating on a
// cycle because each node relays at most once.
//
// Grouped by FORWARDER (whichever node's own hop reached a given recipient
// first), one line per forwarder listing every node it forwarded to and the
// shared delta triple, e.g.:
//   5 → 2, 8 : (dA, dB, dC)
//   2 → 1, 6 : (dA, dB, dC)
//
// All read-only from the content buffer (Node block's per-row GotForwardMsg/
// ForwardDeltaA-C/ForwardFromRow columns), via useDeltaForwardRows (overlay-flags.ts).
// No local state, no domain authoring — pure reflect, mirroring AbcDragLabel's pattern.
export function DeltaForwardLabel() {
  const rows = useDeltaForwardRows();
  const mount = document.getElementById("delta-forward-mount");
  if (!mount) return null;

  // Group by forwarder name: each forwarder gets one line listing every recipient
  // name it forwarded to, plus the shared delta triple (identical for every
  // recipient of the same forwarder, since it's the same message replayed).
  const byForwarder = new Map<string, { names: string[]; dA: number; dB: number; dC: number }>();
  for (const r of rows) {
    const entry = byForwarder.get(r.forwarderName);
    if (entry) {
      entry.names.push(r.name);
    } else {
      byForwarder.set(r.forwarderName, { names: [r.name], dA: r.dA, dB: r.dB, dC: r.dC });
    }
  }

  return createPortal(
    <span className="delta-forward-label">
      <span className="delta-forward-label-header">delta forwarded</span>
      {[...byForwarder.entries()].map(([forwarderName, entry]) => (
        <span className="delta-forward-label-row" key={forwarderName}>
          {forwarderName} → {entry.names.join(", ")} : ({entry.dA}, {entry.dB}, {entry.dC})
        </span>
      ))}
    </span>,
    mount,
  );
}
