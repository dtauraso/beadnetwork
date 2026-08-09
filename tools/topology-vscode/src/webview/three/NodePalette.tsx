// NodePalette.tsx — drag a kind onto the scene to create a node there; press delete to
// remove the selected one (docs/planning/node-palette.md).
//
// RENDER AND FORWARD ONLY. The palette holds no graph: it lists the kinds from NODE_DEFS
// (the single registry) and fires one addressed edit per gesture, fire-and-forget. Where the
// new node connects is not decided here — Go picks the nearest existing node from its own
// geometry, so this sends a POINT and never measures proximity. What is selected is not
// decided here either: Go owns selection, so delete sends the row Go says is selected.
//
// Both edits make Go persist the tree and END THE RUN — a node's stream is a dedicated fd
// the host allocates at spawn, so a node created in-process would have no stream to emit on.
// The editor goes away and comes back with the new tree, the same way a scene-tab switch
// already works. That is why a create does not animate: there is nothing to animate.

import React, { useEffect, useRef, useState } from "react";
import { postGoRecord } from "../vscode-api";
import { encodeSceneCreate, encodeSceneDelete } from "../../schema/input-layout";
import { NODE_KIND_NAMES, NODE_DEFS } from "../../schema/node-defs";
import { useSelectedNodeRow, useSceneEditable, useSceneKinds, useEditRefused } from "./overlay-flags";
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
} from "./overlay-chrome";

// The MIME the drag carries. A private type, not text/plain: a stray text drop from another
// app must not read as a node creation.
const KIND_MIME = "application/x-wirefold-kind";

/** The palette pill: a kind per row, each draggable onto the scene. */
export function NodePalette() {
  const editable = useSceneEditable();
  const sceneKinds = useSceneKinds();
  const [open, setOpen] = useState(false);
  const selectedRow = useSelectedNodeRow();

  // DELETE KEY. Go owns selection, so this forwards the row Go says is selected and nothing
  // else — no local "which node did I click" to disagree with it. Bound while the palette is
  // mounted, i.e. only in a scene that can be edited.
  useEffect(() => {
    if (!editable) return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key !== "Delete" && e.key !== "Backspace") return;
      // Not while typing into something — a text field's delete is its own.
      const t = e.target as HTMLElement | null;
      if (t && (t.tagName === "INPUT" || t.tagName === "TEXTAREA" || t.isContentEditable)) return;
      if (selectedRow < 0) return;
      postGoRecord(encodeSceneDelete(selectedRow));
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [editable, selectedRow]);

  // A scene Go says is not editable renders no palette at all — not a disabled one. The
  // question "can this scene be edited" is Go's (SceneTab.Editable), the same shape every
  // other per-scene property has.
  if (!editable) return null;

  const onToggle = (e: React.MouseEvent) => {
    e.stopPropagation();
    setOpen((o) => !o);
  };

  return (
    <div style={PILL_ANCHOR_STYLE}>
      <div style={pillContainerStyle(false)}>
        {/* No master toggle — there is nothing to turn on or off, only kinds to drag — so
            the WHOLE pill opens the popover, as on the angles and distances pills. The label
            takes the pill's slack so the caret stays at the far end. */}
        <div
          onClick={onToggle}
          title={open ? "Close node palette" : "Open node palette"}
          style={{ ...pillBodyStyle, flex: "1 1 auto" }}
        >
          Nodes
        </div>
        {/* The caret this pill was missing. Every other pill in the column has one, and it is
            the affordance that says the thing opens — without it the pill reads as a button
            that does something rather than a menu. */}
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
          {/* Only the kinds THIS SCENE takes (SceneTab.Kinds, streamed as a kind-id
              bitmask). Offering a kind the scene has no place for and then refusing the drop
              teaches nothing and looks broken; not offering it says the same thing before
              the gesture. Go checks the same mask anyway — the tree is written on that side,
              and "the UI does not offer it" is not "it cannot happen". */}
          {NODE_KIND_NAMES.map((kind, kindId) =>
            sceneKinds & (1 << kindId) ? <PaletteRow key={kind} kind={kind} kindId={kindId} /> : null,
          )}
        </div>
      )}
      <RefusedNotice />
    </div>
  );
}

/** One draggable kind. The kind's OWN id travels in the drag payload, not its name: the id
 *  is what crosses the bridge (the Node block's KindId column already uses it). */
function PaletteRow({ kind, kindId }: { kind: string; kindId: number }) {
  const [hover, setHover] = useState(false);
  // Collapsed by default, like every other disclosure in this chrome: thirteen open
  // descriptions is a wall of prose where a menu should be.
  const [open, setOpen] = useState(false);
  // The heading element itself, handed to the drag as its image — see onDragStart.
  const headingRef = useRef<HTMLSpanElement>(null);
  const def = NODE_DEFS[kind];
  return (
    <div
      draggable
      onDragStart={(e) => {
        e.dataTransfer.setData(KIND_MIME, String(kindId));
        e.dataTransfer.effectAllowed = "copy";
        // THE DRAG CARRIES THE HEADING ONLY — swatch and name. The browser's default ghost is
        // the whole dragged element, which for an open row is the description too: a
        // paragraph floating under the cursor while you aim a drop. The heading is what
        // identifies the kind, and it is the only part worth carrying.
        if (headingRef.current) {
          e.dataTransfer.setDragImage(headingRef.current, 12, 8);
        }
      }}
      onMouseEnter={() => setHover(true)}
      onMouseLeave={() => setHover(false)}
      style={{ ...popoverRowStyle(hover, false), cursor: "grab", flexDirection: "column", alignItems: "stretch", gap: 2 }}
      title={`Drag ${kind} onto the scene`}
    >
      {/* THE HEADING — click the triangle to read what this kind is, drag the row itself to
          make one. Both live on the same row because they are two things you want to do with
          the same kind, and a separate "info" affordance beside every name would be a second
          column of chrome carrying one bit. */}
      <span
        ref={headingRef}
        style={{ display: "flex", alignItems: "center", gap: 7 }}
        onClick={(e) => {
          e.stopPropagation();
          setOpen((o) => !o);
        }}
      >
        <span style={DISCLOSURE_GLYPH_STYLE}>{open ? "▼" : "▶"}</span>
        {/* The kind's own swatch, from its NODE_DEFS entry — the same fill and stroke the
            node itself is drawn with, so the palette shows what you are about to get. */}
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
      {/* THE DESCRIPTION, revealed. Generated from the kind's own SPEC.md — the same file
          that already says what its ports and colours are, so the menu cannot describe a
          kind the code does not.
          REVEALED_LIST_STYLE: it lays out at the popover's width but MEASURES AS NOTHING, so
          a sentence expanding here wraps down the popover instead of stretching it (and the
          pills that share that width) out across the scene. That is what makes a full
          sentence affordable in a menu at all. */}
      {open && def?.desc && (
        <div style={REVEALED_LIST_STYLE}>
          <span
            style={{
              display: "block",
              opacity: 0.8,
              // WRAP, hard. Long prose in a narrow popover has to break wherever it must —
              // `anywhere` breaks mid-word rather than pushing one long token past the edge.
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

/** Says that Go refused a structural edit. The scene looks unchanged after a refusal, which
 *  is exactly why this exists: a gesture that does nothing and says nothing is
 *  indistinguishable from a broken build. The REASON is in the output channel and
 *  .probe/go-errors.jsonl; this only reports that it happened. */
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

/** dropKindFromEvent reads the palette's private MIME off a drop, or null when the drop came
 *  from somewhere else. Exported for the canvas handler that owns the drop target. */
export function dropKindFromEvent(e: DragEvent): number | null {
  const raw = e.dataTransfer?.getData(KIND_MIME);
  if (!raw) return null;
  const id = Number(raw);
  return Number.isInteger(id) && id >= 0 ? id : null;
}

/** Fire the create at a drop's SCREEN position, in normalized device coordinates. Where that
 *  is in the scene is Go's to work out — it has the camera. TS forwards the pointer, the same
 *  as raw-input does, and computes no geometry. */
export function fireCreateAt(kindId: number, ndcX: number, ndcY: number): void {
  postGoRecord(encodeSceneCreate(kindId, ndcX, ndcY));
}
