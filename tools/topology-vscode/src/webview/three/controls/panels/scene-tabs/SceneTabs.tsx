import React, { useCallback } from "react";
import { postGoRecord } from "../../../../vscode-api";
import { encodeSceneSelected } from "../../../../../schema/input/input-encode";
import { postLog } from "../../../../log/post";
import { useSceneTabs } from "./scene-tabs";
import * as T from "../../chrome-theme";

const stripStyle: React.CSSProperties = {

  position: "absolute",
  top: 12,
  left: "50%",
  transform: "translateX(-50%)",
  zIndex: 20,
  pointerEvents: "auto",
  display: "inline-flex",
  flexDirection: "row",
  gap: 2,
  background: T.CHIP,
  border: `1px solid ${T.BORDER}`,
  borderRadius: T.RADIUS_CHIP,
  padding: T.PAD_CHIP,
  fontSize: T.FONT_SIZE,
  fontFamily: T.FONT_STACK,
  userSelect: "none",
};

function tabStyle(active: boolean): React.CSSProperties {
  return {
    background: active ? T.ACCENT : "transparent",
    border: "none",
    borderRadius: T.RADIUS_ITEM,
    color: active ? T.ACCENT_INK : T.TEXT,
    fontSize: T.FONT_SIZE,
    fontFamily: T.FONT_STACK,
    lineHeight: 1,
    padding: T.PAD_CHIP,
    cursor: "pointer",
  };
}

export function SceneTabs() {
  const { names, selected } = useSceneTabs();
  const onPick = useCallback((index: number) => {
    postLog("scene-tab-click", { index });
    postGoRecord(encodeSceneSelected(index));
  }, []);

  if (names.length === 0) return null;

  return (
    <div style={stripStyle}>
      {names.map((name, i) => (
        <button
          key={name}
          type="button"
          style={tabStyle(i === selected)}
          title={i === selected ? `showing "${name}"` : `switch to "${name}" (reloads the sim)`}
          onClick={(e) => {
            e.stopPropagation();
            onPick(i);
          }}
        >
          {name}
        </button>
      ))}
    </div>
  );
}
