import type React from "react";
import { createPortal } from "react-dom";
import { postGoRecord } from "../../../vscode-api";
import { encodeClockSpeed } from "../../../../schema/input-encode";
import { usePlaybackSpeed } from "../flags/overlay-flags-speed";
import { SPEED_SETTINGS, settingKey, DEFAULT_INDEX, closestSettingIndex } from "./speed-settings";

// A fraction carries `num`/`den` and is drawn stacked over a bar (see fracStyle); a whole
// number carries `label` and is drawn as itself. The quarters were previously the single
// precomposed glyphs ¼/½/¾, which render stacked-over-a-bar in this font — but with their
// numerator and denominator almost touching, and a glyph's internal spacing cannot be
// adjusted. Composing the same two digits, AT THE SAME SIZE the glyph draws them
// (FRAC_EM), is what makes that one gap settable. This is a gap change only: nothing here
// resizes a component or restyles it.
//
// The table itself (SPEED_SETTINGS/settingKey/DEFAULT_INDEX/closestSettingIndex) lives in
// speed-settings.ts — no react/vscode-api dependency, same split as tilt-vector-angle-format.ts.

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
    <span className="speed-slider" style={namedWrapStyle}>
      {/* The slider's own name, at the head of the toolbar — the spot a static "saved"
          label used to occupy (html.ts). A control on a bar with other controls has to say
          which one it is. */}
      <span style={nameStyle}>speed</span>
      <span style={wrapStyle}>
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
            {"label" in setting ? (
              setting.label
            ) : (
              <span style={fracStyle}>
                <span>{setting.num}</span>
                <span style={fracBarStyle} />
                <span>{setting.den}</span>
              </span>
            )}
          </span>
        ))}
        </span>
      </span>
    </span>,
    mount,
  );
}

// The name sits BESIDE the control; the control itself is the column below.
const namedWrapStyle: React.CSSProperties = {
  display: "inline-flex",
  flexDirection: "row",
  alignItems: "center",
  gap: 8,
};

// Baseline-ish with the slider rather than with its tick row, so the name reads against the
// track and not against the fractions under it.
const nameStyle: React.CSSProperties = {
  color: "#333",
  whiteSpace: "nowrap",
};

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
// not in one of the dark overlay panels floating over the canvas — the dark-panel palette
// (#ddd text) is near-white on white and left them all but unreadable.
//
// Every label is full-strength black. Nothing here is dimmed to push it back: a label that
// is shown is shown to be read, and greying the unselected ones makes five of the six
// settings harder to read in exchange for saying something WEIGHT already says.
const tickStyle: React.CSSProperties = { color: "#000" };

// --- the stacked fraction ---
//
// FRAC_EM is the numerator/denominator size as a fraction of the row's own font size. It is
// set to what the precomposed ¼/½/¾ glyphs draw their own digits at, so replacing a glyph
// with two composed digits does not change how big anything looks — only the gap between
// them moves.
const FRAC_EM = 0.62;

// FRAC_GAP is the ONE number this change exists to set: the space above and below the bar,
// in px. The glyphs sit their numerator and denominator almost against the bar; this opens
// that up by roughly 2px total.
const FRAC_GAP = 1;

const fracStyle: React.CSSProperties = {
  display: "inline-flex",
  flexDirection: "column",
  alignItems: "center",
  justifyContent: "center",
  verticalAlign: "middle",
  fontSize: `${FRAC_EM}em`,
  lineHeight: 1,
};

// The fraction bar, drawn in the text's own colour so it bolds along with the digits when
// this is the selected position.
const fracBarStyle: React.CSSProperties = {
  display: "block",
  width: "100%",
  height: 1,
  margin: `${FRAC_GAP}px 0`,
  background: "currentColor",
};

// The selected position is marked by WEIGHT ALONE — same colour as the rest, just bold.
const tickOnStyle: React.CSSProperties = { color: "#000", fontWeight: "bold" };
