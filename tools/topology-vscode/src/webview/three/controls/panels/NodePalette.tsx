import React, { useEffect, useState } from "react";
import { postGoRecord } from "../../../vscode-api";
import { encodeSceneDelete } from "../../../../schema/input-encode-scene-tilt";
import { NODE_KIND_NAMES } from "../../../../schema/node-defs";
import { useSelectedNodeRow } from "../flags/overlay-flags-selection";
import { useSceneEditable, useSceneKinds } from "../flags/overlay-flags-scene";
import { useEditRefused } from "../flags/overlay-flags-edit-refused";
import { PaletteRow } from "./PaletteRow";
import {
  pillContainerStyle,
  pillBodyStyle,
  pillCaretStyle,
  PILL_ANCHOR_STYLE,
  inFlowPopoverStyle,
  CHROME_TEXT,
  CHROME_FONT_STACK,
} from "../pills/overlay-chrome";

export function NodePalette() {
  const editable = useSceneEditable();
  const sceneKinds = useSceneKinds();
  const [open, setOpen] = useState(false);
  const selectedRow = useSelectedNodeRow();

  useEffect(() => {
    if (!editable) return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key !== "Delete" && e.key !== "Backspace") return;

      const t = e.target as HTMLElement | null;
      if (t && (t.tagName === "INPUT" || t.tagName === "TEXTAREA" || t.isContentEditable)) return;
      if (selectedRow < 0) return;
      postGoRecord(encodeSceneDelete(selectedRow));
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [editable, selectedRow]);

  if (!editable) return null;

  const onToggle = (e: React.MouseEvent) => {
    e.stopPropagation();
    setOpen((o) => !o);
  };

  return (
    <div style={PILL_ANCHOR_STYLE}>
      <div style={pillContainerStyle(false)}>
        {}
        <div
          onClick={onToggle}
          title={open ? "Close node palette" : "Open node palette"}
          style={{ ...pillBodyStyle, flex: "1 1 auto" }}
        >
          Nodes
        </div>
        {}
        <div
          onClick={onToggle}
          title={open ? "Close node palette" : "Open node palette"}
          style={pillCaretStyle}
        >
          {open ? "▲" : "▼"}
        </div>
      </div>
      {open && (
        <div style={inFlowPopoverStyle()}>
          {}
          {NODE_KIND_NAMES.map((kind, kindId) =>
            sceneKinds & (1 << kindId) ? <PaletteRow key={kind} kind={kind} kindId={kindId} /> : null,
          )}
        </div>
      )}
      <RefusedNotice />
    </div>
  );
}

function RefusedNotice() {
  const refused = useEditRefused();
  const [seen, setSeen] = useState(refused);
  const [showing, setShowing] = useState(false);

  useEffect(() => {
    if (refused === seen) return;
    setSeen(refused);
    setShowing(true);
    const t = window.setTimeout(() => setShowing(false), 4000);
    return () => window.clearTimeout(t);
  }, [refused, seen]);

  if (!showing) return null;
  return (
    <div
      style={{
        pointerEvents: "auto",
        background: "#34343d",
        border: "1px solid #3a3a44",
        borderRadius: 6,
        color: CHROME_TEXT,
        fontFamily: CHROME_FONT_STACK,
        fontSize: 11,
        padding: "3px 8px",
      }}
    >
      edit refused — see the output channel
    </div>
  );
}
