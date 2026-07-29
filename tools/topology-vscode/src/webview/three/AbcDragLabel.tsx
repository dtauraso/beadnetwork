import { createPortal } from "react-dom";
import { useAbcDragRows, useDraggedNodeName, useDraggedNodeRelay } from "./overlay-flags";

// AbcDragLabel — the in-editor "drag received" log. A header line carrying the
// LAST-dragged node's own name (persists past pointerup — see DragNodeRow's latch,
// nodes/Wiring ui_state.go lastDraggedNode) followed by that node's CASCADE RELAY word
// in parentheses — fan / routed / terminus, the Node block's CascadeRelay column,
// which says whether a delta triple this node picks up goes on to every cascade
// neighbor, to one sender-chosen kind, or nowhere (Go decides it: node_mover.go's
// cascadeRelayClass). Then a "drag received" label, then ONE
// LINE PER RECIPIENT: that node's name, ITS OWN per-node received count, and the
// delta triple (dA,dB,dC) it received. The delta is the DRAGGED node's own
// quantized-triple change, computed once at the drag and carried on the
// neighborSetC message — a recipient reports the delta it was handed, it does not
// apply it. The count is that SAME recipient's own cumulative DragRequantCount, not
// a summed total — there is no single "how many total" number anymore, only each
// recipient's own tally.
//
// All read-only from the content buffer (Overlay block's DragNodeRow column, plus the
// Node block's per-row Label section and GotDragMsg/DragDeltaA/B/C/DragRequantCount
// columns), via the buffer-reflect hooks in overlay-flags.ts. Go owns dragged-node
// identity as the gesture FSM's g.dragNode (nodes/Wiring/gesture.go), latched into
// uiState.lastDraggedNode at drag-start so it survives the drag ending; DragNodeRow
// carries its ROW INDEX, and the name is resolved from that row's own Label — there
// is no name/id sidecar. Go-owned and drag-scoped: Go clears the recipient set AND
// each recipient's own count at drag start (KindAbcDragReset → resetAbcDrag) and
// emits the cleared state, so an empty list (and a zeroed count) is meaningful and
// must render.
//
// No local state, no domain authoring. Mirrors SpeedSlider's portal-into-toolbar-mount
// pattern, just reading instead of writing.
export function AbcDragLabel() {
  const draggedName = useDraggedNodeName();
  const relay = useDraggedNodeRelay();
  const rows = useAbcDragRows();
  const mount = document.getElementById("abc-drag-mount");
  if (!mount) return null;

  return createPortal(
    <span className="abc-drag-label">
      {draggedName && (
        <span className="abc-drag-label-header">
          dragging {draggedName}
          {relay && <span className="abc-drag-label-relay"> ({relay})</span>}
        </span>
      )}
      <span className="abc-drag-label-header">drag received</span>
      {rows.map((r) => (
        <span className="abc-drag-label-row" key={r.name}>
          {r.name}: ×{r.count} ({r.dA}, {r.dB}, {r.dC})
        </span>
      ))}
    </span>,
    mount,
  );
}
