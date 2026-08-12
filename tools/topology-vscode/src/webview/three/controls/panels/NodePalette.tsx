













import React, { useEffect, useRef, useState } from "react";
import { postGoRecord } from "../../../vscode-api";
import { encodeSceneCreate, encodeSceneDelete } from "../../../../schema/input-encode";
import { NODE_KIND_NAMES, NODE_DEFS } from "../../../../schema/node-defs";
import { useSelectedNodeRow } from "../flags/overlay-flags-selection";
import { useSceneEditable, useSceneKinds } from "../flags/overlay-flags-scene";
import { useEditRefused } from "../flags/overlay-flags-edit-refused";
import {
  pillContainerStyle,
  pillBodyStyle,
  pillCaretStyle,
  PILL_ANCHOR_STYLE,
  inFlowPopoverStyle,
  popoverRowStyle,
  CHROME_TEXT,
  CHROME_FONT_STACK,
  DISCLOSURE_GLYPH_STYLE,
  REVEALED_LIST_STYLE,
} from "../pills/overlay-chrome";



const KIND_MIME = "application/x-wirefold-kind";


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


function PaletteRow({ kind, kindId }: { kind: string; kindId: number }) {
  const [hover, setHover] = useState(false);


  const [open, setOpen] = useState(false);

  const headingRef = useRef<HTMLSpanElement>(null);
  const def = NODE_DEFS[kind];
  return (
    <div
      draggable
      onDragStart={(e) => {
        e.dataTransfer.setData(KIND_MIME, String(kindId));
        e.dataTransfer.effectAllowed = "copy";




        if (headingRef.current) {
          e.dataTransfer.setDragImage(headingRef.current, 12, 8);
        }
      }}
      onMouseEnter={() => setHover(true)}
      onMouseLeave={() => setHover(false)}
      style={{ ...popoverRowStyle(hover, false), cursor: "grab", flexDirection: "column", alignItems: "stretch", gap: 2 }}
      title={`Drag ${kind} onto the scene`}
    >
      {}
      <span
        ref={headingRef}
        style={{ display: "flex", alignItems: "center", gap: 7 }}
        onClick={(e) => {
          e.stopPropagation();
          setOpen((o) => !o);
        }}
      >
        <span style={DISCLOSURE_GLYPH_STYLE}>{open ? "▼" : "▶"}</span>
        {}
        <span
          style={{
            width: 11,
            height: 11,
            flex: "0 0 auto",
            borderRadius: 3,
            background: def?.fill ?? "#888",
            border: `1px solid ${def?.stroke ?? "#888"}`,
          }}
        />
        <span style={{ minWidth: 0, overflowWrap: "anywhere" }}>{kind}</span>
      </span>
      {}
      {open && def?.desc && (
        <div style={REVEALED_LIST_STYLE}>
          <span
            style={{
              display: "block",
              opacity: 0.8,


              overflowWrap: "anywhere",
              whiteSpace: "normal",
              lineHeight: 1.35,
              paddingLeft: 18,
            }}
          >
            {def.desc}
          </span>
        </div>
      )}
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


export function dropKindFromEvent(e: DragEvent): number | null {
  const raw = e.dataTransfer?.getData(KIND_MIME);
  if (!raw) return null;
  const id = Number(raw);
  return Number.isInteger(id) && id >= 0 ? id : null;
}


export function fireCreateAt(kindId: number, ndcX: number, ndcY: number): void {
  postGoRecord(encodeSceneCreate(kindId, ndcX, ndcY));
}
