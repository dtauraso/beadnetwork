import type React from "react";
import { createPortal } from "react-dom";
import { postGoRecord } from "../../../../vscode-api";
import { encodeClockSpeed } from "../../../../../schema/input-encode";
import { usePlaybackSpeed } from "../../flags/overlay-flags-speed";
import { SPEED_SETTINGS, settingKey, DEFAULT_INDEX, closestSettingIndex } from "./speed-settings";

const TRACK_W = 104;

const THUMB_INSET = 6;

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
      {}
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
      {}
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

const namedWrapStyle: React.CSSProperties = {
  display: "inline-flex",
  flexDirection: "row",
  alignItems: "center",
  gap: 8,
};

const nameStyle: React.CSSProperties = {
  color: "#333",
  whiteSpace: "nowrap",
};

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

const tickStyle: React.CSSProperties = { color: "#000" };

const FRAC_EM = 0.62;

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

const fracBarStyle: React.CSSProperties = {
  display: "block",
  width: "100%",
  height: 1,
  margin: `${FRAC_GAP}px 0`,
  background: "currentColor",
};

const tickOnStyle: React.CSSProperties = { color: "#000", fontWeight: "bold" };
