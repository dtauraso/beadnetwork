import { createPortal } from "react-dom";
import { postGoRecord } from "../vscode-api";
import { encodeClockSpeed } from "../../schema/input-layout";
import { usePlaybackSpeed } from "./overlay-flags";

// SPEED_TABLE — the six settings the slider's raw <input> value (0..5) indexes into. The
// slider position is an INDEX, never the multiplier itself; the multiplier is what crosses
// the wire and what the label displays.
const SPEED_TABLE = [0, 0.25, 0.5, 0.75, 1, 2] as const;

// SpeedSlider — a playback-speed control. Speed is Go-owned state (the clock), persisted
// by Go to view/speed.json and streamed back on the VIEW frame's Overlay.Speed column; this
// component holds NO local speed state (no store — memory/feedback_reflect_dont_create_store.md)
// — it reflects the buffer via usePlaybackSpeed and fire-and-forgets each change to Go via
// the edit-update(clock, speed) wire record. No await, no Promise chain
// (check-no-await-on-bridge).
export function SpeedSlider() {
  const speed = usePlaybackSpeed();
  const mount = document.getElementById("run-mount");
  if (!mount) return null;

  // Default to index 4 (multiplier 1) until the first snapshot decodes, matching Go's own
  // defaultPlaybackSpeed fallback for a missing/malformed speed.json.
  const index = speed == null ? 4 : Math.max(0, SPEED_TABLE.indexOf(closestTableValue(speed)));

  const onChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const i = Number(e.target.value);
    const value = SPEED_TABLE[i] ?? 1;
    postGoRecord(encodeClockSpeed(value));
  };

  return createPortal(
    <span className="speed-slider">
      <input
        type="range"
        min={0}
        max={SPEED_TABLE.length - 1}
        step={1}
        value={index}
        onChange={onChange}
        aria-label="playback speed"
      />
      <span className="speed-slider-label">{SPEED_TABLE[index]}</span>
    </span>,
    mount,
  );
}

// closestTableValue finds the SPEED_TABLE entry nearest a decoded float32 speed (a float32
// round-trip of e.g. 0.25 is exact, but this guards against any future non-table value
// reaching the buffer, rather than indexOf silently failing to -1).
function closestTableValue(speed: number): (typeof SPEED_TABLE)[number] {
  let best: (typeof SPEED_TABLE)[number] = SPEED_TABLE[0];
  let bestDiff = Infinity;
  for (const v of SPEED_TABLE) {
    const diff = Math.abs(v - speed);
    if (diff < bestDiff) {
      bestDiff = diff;
      best = v;
    }
  }
  return best;
}
