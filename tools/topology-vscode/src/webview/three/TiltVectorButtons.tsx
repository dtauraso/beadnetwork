import { createPortal } from "react-dom";
import { postGoRecord } from "../vscode-api";
import { encodeTiltVectorStart, encodeTiltVectorReset } from "../../schema/input-layout";
import { useTiltVectorRows } from "./overlay-flags";

// TiltVectorButtons — the START TILT and RESET TILT controls, portaled into "#run-mount",
// the SAME toolbar mount SpeedSlider uses (html.ts). #run-mount itself lays out its direct
// children in a COLUMN (webview.css), so the two buttons are rendered together as ONE row
// span here — START first, then RESET, JSX order left-to-right in a row — rather than as
// two separate top-level portals, which would stack them as two more column rows instead of
// placing START immediately to the left of RESET.
//
// WHICH nodes these can act on is the SAME data-driven signal both controls have always
// used: useTiltVectorRows (overlay-flags.ts), which reflects every node whose
// TopTiltVectorLen > 0. A scene with no tilt vectors (e.g. the ring) yields an empty row
// list and this renders nothing — no scene-name check anywhere in TS.
//
// Each click fire-and-forgets one edit-update(tiltVector, start|reset) record PER row
// currently shown, each naming that node's buffer ROW (never its id/name — no sidecar). Go
// owns what each means and applies it on that node's own goroutine
// (nodes/Node1/node.go / nodes/Node2/node.go's applyTiltEdit for the pair,
// node_mover.go's moveMsgKindTiltVectorReset for every other kind's reset — start has no
// mover fallback, since it is meaningless off the pair's own vector exchange):
//
//   - START opens the vector exchange from whatever angles are currently set — this is
//     "the kick" a ▲/▼ panel click used to fire as a side effect (task/pair-node-owns-itself
//     split, so a panel click now moves the tilt by exactly one π/12 step and nothing else).
//   - RESET is a stop-and-return: both tilt-angle indices back to 0 (world +y), no bead
//     placed, and the straightening exchange never starts.
export function TiltVectorButtons() {
  const rows = useTiltVectorRows();
  const mount = document.getElementById("run-mount");
  if (!mount || !rows || rows.length === 0) return null;

  const start = () => {
    for (const row of rows) {
      postGoRecord(encodeTiltVectorStart(row.row));
    }
  };
  const reset = () => {
    for (const row of rows) {
      postGoRecord(encodeTiltVectorReset(row.row));
    }
  };

  return createPortal(
    <span className="tilt-vector-btn-row">
      <button type="button" className="run-btn tilt-start-btn" onClick={start} aria-label="start tilt exchange">
        start tilt
      </button>
      <button type="button" className="run-btn tilt-reset-btn" onClick={reset} aria-label="reset tilt vectors">
        reset tilt
      </button>
    </span>,
    mount,
  );
}
