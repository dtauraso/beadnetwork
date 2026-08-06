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
// The quarters read as FRACTIONS rather than decimals because the table IS quarters of the
// clock (nodes/wire.MsPerTick is divisible by 4 so they divide exactly —
// clock_ms_per_tick_quarters_test.go pins that); "1/4" says that where "0.25" reads as an
// arbitrary decimal.
//
// A fraction carries `num`/`den` as separate strings and is BUILT from three spans, rather
// than being the single precomposed glyph "¼". At the 11px this row renders at, that glyph's
// own numerator and denominator are a couple of pixels tall and sit almost on top of each
// other — legible as "a fraction", not as WHICH fraction. Built from parts, the numerator
// and denominator can be pushed apart around the slash (see numStyle/denStyle) so each digit
// is readable at that size. A whole number has no `den` and renders as a plain label.
const SPEED_SETTINGS = [
  { speed: 0, num: "0" },
  { speed: 0.25, num: "1", den: "4" },
  { speed: 0.5, num: "1", den: "2" },
  { speed: 0.75, num: "3", den: "4" },
  { speed: 1, num: "1" },
  { speed: 2, num: "2" },
] as const;

// key identifies a setting for React's list reconciliation and for the aria label — the
// speed itself, which is unique across the table by construction.
const settingKey = (s: (typeof SPEED_SETTINGS)[number]): string => String(s.speed);

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
          <span key={settingKey(setting)} style={i === index ? tickOnStyle : tickStyle}>
            <span style={numStyle}>{setting.num}</span>
            {"den" in setting && (
              <>
                <span style={slashStyle}>/</span>
                <span style={denStyle}>{setting.den}</span>
              </>
            )}
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
  fontSize: 11,
  fontFamily: "monospace",
  lineHeight: 1,
  userSelect: "none",
  pointerEvents: "none",
};

// These labels sit in the LIGHT toolbar (.toolbar in webview.css is `background: #fff`),
// not in one of the dark overlay panels floating over the canvas. They are coloured from
// the toolbar's own palette — the same #555/#333 the other toolbar-adjacent labels
// (.abc-drag-label, .rule-eq-panel) use — because the dark-panel palette (#ddd, and a 0.5
// opacity dim on top of it) is near-white on white and left them all but unreadable.
const tickStyle: React.CSSProperties = { color: "#555" };

// --- fraction parts ---
//
// The numerator rides above the slash and the denominator below it, pushed apart by
// FRAC_SHIFT so the two digits read separately instead of colliding the way the
// precomposed "¼" glyph does at this size. translateY moves them WITHOUT changing the
// line's own height (unlike vertical-align, which grows the line box and would shove the
// label row down away from the track), so the tick row stays tight under the slider.
const FRAC_SHIFT = 2;

const numStyle: React.CSSProperties = {
  display: "inline-block",
  transform: `translateY(-${FRAC_SHIFT}px)`,
};

const denStyle: React.CSSProperties = {
  display: "inline-block",
  transform: `translateY(${FRAC_SHIFT}px)`,
};

// The slash sits between them, styled exactly like the digits — same colour, same weight,
// same size. Only the vertical offset differs across the three parts; anything else (a
// dimmed slash, a tightened margin) restyles the label rather than separating it.
const slashStyle: React.CSSProperties = {
  display: "inline-block",
};

// The selected position: darkest and bold, so which setting is live is read off the same
// row that shows what the settings are.
const tickOnStyle: React.CSSProperties = { color: "#000", fontWeight: "bold" };
