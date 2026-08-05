import { createPortal } from "react-dom";
import { postGoRecord } from "../vscode-api";
import { encodeTiltVectorReset } from "../../schema/input-layout";
import { useTiltVectorRows } from "./overlay-flags";

// TiltResetButton — a single RESET control, rendered BELOW the speed bar (portaled into
// its own mount, "#tilt-reset-mount", the same createPortal-outside-#app pattern as
// SpeedSlider — see html.ts) rather than into the 3D overlay column
// TiltVectorAnglePanel occupies.
//
// WHICH nodes it can reset is the SAME data-driven signal TiltVectorAnglePanel already
// uses: useTiltVectorRows (overlay-flags.ts), which reflects every node whose
// TopTiltVectorLen > 0. A scene with no tilt vectors (e.g. the ring) yields an empty row
// list and this button renders nothing — no scene-name check anywhere in TS.
//
// A click fire-and-forgets one edit-update(tiltVector, reset) record PER row currently
// shown, each naming that node's buffer ROW (never its id/name — no sidecar). Go owns
// what "start position" means (both tilt-angle indices back to 0, i.e. pointing at world
// +y) and applies it on that node's own goroutine
// (nodes/Node1/node.go / nodes/Node2/node.go's applyTiltEdit for the pair,
// node_mover.go's moveMsgKindTiltVectorReset for every other kind). This is a
// stop-and-return: unlike a panel angle click ("the kick"), no bead is placed and the
// straightening exchange never starts.
export function TiltResetButton() {
  const rows = useTiltVectorRows();
  const mount = document.getElementById("tilt-reset-mount");
  if (!mount || !rows || rows.length === 0) return null;

  const reset = () => {
    for (const row of rows) {
      postGoRecord(encodeTiltVectorReset(row.row));
    }
  };

  return createPortal(
    <button type="button" className="run-btn tilt-reset-btn" onClick={reset} aria-label="reset tilt vectors">
      reset tilt
    </button>,
    mount,
  );
}
