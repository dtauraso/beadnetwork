import React, { useCallback } from "react";
import { postGoRecord } from "../../../../vscode-api";
import { encodeSceneSelected } from "../../../../../schema/input-encode";
import { postLog } from "../../../../log/post";
import { useSceneTabs } from "./scene-tabs";

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
  background: "rgba(0,0,0,0.55)",
  borderRadius: 6,
  padding: "3px 4px",
  fontSize: 11,
  fontFamily: "monospace",
  userSelect: "none",
};

function tabStyle(active: boolean): React.CSSProperties {
  return {
    background: active ? "rgba(255,255,255,0.22)" : "transparent",
    border: "none",
    borderRadius: 4,
    color: active ? "#fff" : "#bbb",
    fontSize: 11,
    fontFamily: "monospace",
    lineHeight: 1,
    padding: "3px 8px",
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
