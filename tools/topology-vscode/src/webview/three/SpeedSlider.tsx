import type React from "react";
import { createPortal } from "react-dom";
import { postGoRecord } from "../vscode-api";
import { encodeClockSpeed } from "../../schema/input-layout";
import { usePlaybackSpeed } from "./overlay-flags";

// SPEED_SETTINGS — the six settings the slider's raw <input> value (0..5) indexes into.
// The slider position is an INDEX, never the multiplier itself: `speed` is what crosses the
// wire, `label` is how that position reads on screen.
//
// One array of pairs rather than a value table beside a label table: two parallel arrays
// can lose step with each other (a setting added to one and not the other), and nothing
// would catch it — the labels would silently name the wrong speeds from that index on.
//
// The quarters read as vulgar fractions rather than decimals because the table IS quarters
// of the clock (nodes/wire.MsPerTick is divisible by 4 so they divide exactly —
// clock_ms_per_tick_quarters_test.go pins that); "¼" says that where "0.25" reads as an
// arbitrary decimal.
const SPEED_SETTINGS = [
  { speed: 0, label: "0" },
  { speed: 0.25, label: "¼" },
  { speed: 0.5, label: "½" },
  { speed: 0.75, label: "¾" },
  { speed: 1, label: "1" },
  { speed: 2, label: "2" },
] as const;

// DEFAULT_INDEX is the position holding multiplier 1 — what the slider shows until the
// first snapshot decodes, matching Go's own defaultPlaybackSpeed fallback for a missing or
// malformed speed.json. Found by value rather than written as a literal 4, so reordering
// or inserting a setting cannot leave it pointing at the wrong one.
const DEFAULT_INDEX = SPEED_SETTINGS.findIndex((s) => s.speed === 1);

// Slider track width, in px, shared by the input and the tick-label row below it so the
// labels line up with the positions they name.
const TRACK_W = 104;

// Half a range thumb, in px. The thumb's CENTRE at the extremes sits this far inside the
// track's own ends, so the label row is inset by the same amount — otherwise "0" and "2"
// sit outboard of the positions they label while the middle four look right.
const THUMB_INSET = 6;

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

  const index = speed == null ? DEFAULT_INDEX : closestSettingIndex(speed);

  const onChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const i = Number(e.target.value);
    postGoRecord(encodeClockSpeed(SPEED_SETTINGS[i]?.speed ?? 1));
  };

  return createPortal(
    <span className="speed-slider" style={wrapStyle}>
      <input
        type="range"
        min={0}
        max={SPEED_SETTINGS.length - 1}
        step={1}
        value={index}
        onChange={onChange}
        aria-label="playback speed"
        style={inputStyle}
      />
      {/* One label per position, under the position it names. The row is the same width as
          the track (inset by half a thumb, see THUMB_INSET) and spreads its labels evenly,
          which puts each one under its own tick because the positions are evenly spaced.
          The current setting is highlighted here rather than printed again beside the
          slider — with every value visible, a separate read-out of the selected one is a
          second place saying the same thing. */}
      <span style={ticksStyle} aria-hidden="true">
        {SPEED_SETTINGS.map((setting, i) => (
          <span key={setting.label} style={i === index ? tickOnStyle : tickStyle}>
            {setting.label}
          </span>
        ))}
      </span>
    </span>,
    mount,
  );
}

// closestSettingIndex finds the SPEED_SETTINGS position whose speed is nearest a decoded
// float32 (a float32 round-trip of e.g. 0.25 is exact, but this guards against any future
// non-table value reaching the buffer, rather than an exact lookup silently missing).
function closestSettingIndex(speed: number): number {
  let best = 0;
  let bestDiff = Infinity;
  SPEED_SETTINGS.forEach((setting, i) => {
    const diff = Math.abs(setting.speed - speed);
    if (diff < bestDiff) {
      bestDiff = diff;
      best = i;
    }
  });
  return best;
}

// The slider and its tick-label row stack, so the labels sit BELOW the positions rather
// than beside the control. inline-flex keeps the whole thing inline in the run-mount
// toolbar, next to whatever else is portaled there.
const wrapStyle: React.CSSProperties = {
  display: "inline-flex",
  flexDirection: "column",
  alignItems: "stretch",
  verticalAlign: "middle",
};

const inputStyle: React.CSSProperties = {
  width: TRACK_W,
  margin: 0,
  display: "block",
};

const ticksStyle: React.CSSProperties = {
  display: "flex",
  justifyContent: "space-between",
  width: TRACK_W,
  padding: `0 ${THUMB_INSET}px`,
  boxSizing: "border-box",
  fontSize: 9,
  fontFamily: "monospace",
  lineHeight: 1,
  userSelect: "none",
  pointerEvents: "none",
};

const tickStyle: React.CSSProperties = { color: "#ddd", opacity: 0.5 };

// The selected position: full-strength and bold, so which setting is live is read off the
// same row that shows what the settings are.
const tickOnStyle: React.CSSProperties = { color: "#fff", opacity: 1, fontWeight: "bold" };
