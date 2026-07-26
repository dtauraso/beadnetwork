import { createPortal } from "react-dom";
import { useDragReceivedCount, useAbcDragRows, useDraggedNodeName } from "./overlay-flags";

// AbcDragLabel — the in-editor "drag received" log. A header carrying the DRAGGED
// node's own name (when a drag is in progress) plus a live count of abc-drag events,
// then ONE LINE PER RECIPIENT: that node's name and the delta triple (dA,dB,dC) it
// received. The delta is the DRAGGED node's own quantized-triple change, computed
// once at the drag and carried on the neighborSetC message — a recipient reports the
// delta it was handed, it does not apply it.
//
// All read-only from the content buffer (Overlay block's DragNodeRow column, plus the
// Node block's per-row Label section and GotDragMsg/DragDeltaA/B/C/DragRequantCount
// columns), via the buffer-reflect hooks in overlay-flags.ts. Go owns dragged-node
// identity as the gesture FSM's g.dragNode (nodes/Wiring/gesture.go); DragNodeRow
// carries its ROW INDEX, and the name is resolved from that row's own Label — there
// is no name/id sidecar. The header count SUMS each recipient's own DragRequantCount
// (no central accumulator — that used to drop ticks under a fast drag's pointer-input
// load; see overlay-flags.ts readDragReceivedCount). Go-owned and drag-scoped: Go
// clears the recipient set AND each recipient's own count at drag start
// (KindAbcDragReset → resetAbcDrag) and emits the cleared state, so an empty list
// (and a zeroed count) is meaningful and must render.
//
// No local state, no domain authoring. Mirrors SpeedSlider's portal-into-toolbar-mount
// pattern, just reading instead of writing.
export function AbcDragLabel() {
  const draggedName = useDraggedNodeName();
  const count = useDragReceivedCount();
  const rows = useAbcDragRows();
  const mount = document.getElementById("abc-drag-mount");
  if (!mount) return null;

  return createPortal(
    <span className="abc-drag-label">
      {draggedName && (
        <span className="abc-drag-label-header">dragging {draggedName}</span>
      )}
      <span className="abc-drag-label-header">drag received ×{count}</span>
      {rows.map((r) => (
        <span className="abc-drag-label-row" key={r.name}>
          {r.name}: ({r.dA}, {r.dB}, {r.dC})
        </span>
      ))}
    </span>,
    mount,
  );
}
