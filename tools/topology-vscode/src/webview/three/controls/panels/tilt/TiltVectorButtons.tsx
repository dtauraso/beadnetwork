import { createPortal } from "react-dom";
import { postGoRecord } from "../../../../vscode-api";
import { encodeTiltVectorStart, encodeTiltVectorReset } from "../../../../../schema/input/input-encode-scene-tilt";
import { useTiltVectorRows } from "../../flags/overlay-flags-tilt-vectors";

export function TiltVectorButtons() {
  const rows = useTiltVectorRows();
  const mount = document.getElementById("tilt-mount");
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
    <>
      <span className="tilt-vector-btn-row">
        <button type="button" className="run-btn tilt-start-btn" onClick={start} aria-label="start tilt exchange">
          start tilt
        </button>
        <button type="button" className="run-btn tilt-reset-btn" onClick={reset} aria-label="reset tilt vectors">
          reset tilt
        </button>
      </span>
      <span
        className="tilt-rounds-readout"
        aria-label="rounds and messages to parallel"
        style={{ gridTemplateColumns: `repeat(${rows.length}, auto)` }}
      >
        {rows.map((row) => (
          <span key={row.row} className="tilt-rounds-node">node {row.label}</span>
        ))}
        {rows.map((row) => (
          <span key={row.row} className="tilt-rounds-cell">
            <span className="tilt-rounds-key">rounds</span>
            <span className="tilt-rounds-val">{row.roundsToParallel}</span>
          </span>
        ))}
        {rows.map((row) => (
          <span key={row.row} className="tilt-rounds-cell">
            <span className="tilt-rounds-key">msgs</span>
            <span className="tilt-rounds-val">{row.msgsToParallel}</span>
          </span>
        ))}
      </span>
    </>,
    mount,
  );
}
