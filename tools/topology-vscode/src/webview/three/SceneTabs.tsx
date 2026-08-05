// SceneTabs.tsx — the scene tab strip: which diagram is showing, and a click to show
// another one.
//
// Everything about a tab is GO-OWNED (nodes/Wiring/scene_tabs.go): the list, the labels,
// which one is selected, what each one loads, and how the switch is performed. This file
// forwards a click as ONE addressed edit (edit-update kind="scene" attr="selected",
// carrying the tab INDEX — no name, no path) and renders what scene-tabs.ts reflects. Same
// shape as the overlay toggles: camera-ui.tsx draws, overlay-flags.ts reflects.
//
// Deliberately NO optimistic highlight. Clicking a tab does not move the highlight here —
// the highlight follows the buffer, so it moves when the newly-loaded scene's first VIEW
// frame arrives. An optimistic one would be TS authoring the selection, and it would lie
// for the whole reload if the switch failed (a write error leaves Go on the old scene,
// scene_tabs.go's SelectScene).

import React, { useCallback } from "react";
import { postGoRecord } from "../vscode-api";
import { encodeSceneSelected } from "../../schema/input-layout";
import { postLog } from "../log/post";
import { useSceneTabs } from "./scene-tabs";

const stripStyle: React.CSSProperties = {
  // Same absolute-positioning scheme and containing block as HomeButton (top 44) /
  // DistanceHomePanel (top 66) / OverlaysControl (top 128) — see DistanceHomePanel's
  // comment. Top-LEFT, since the right column is taken.
  position: "absolute",
  top: 12,
  left: 12,
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

  // No tabs = untabbed anchor (or no frame yet): render nothing at all rather than an
  // empty chrome box.
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
