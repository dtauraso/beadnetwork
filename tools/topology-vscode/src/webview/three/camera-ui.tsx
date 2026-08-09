// camera-ui.tsx — standalone camera control UI widgets for ThreeView.
// OverlaysControl (split-button + popover), HomeButton — no scene/Go logic.

import React, { useCallback, useState } from "react";
import * as THREE from "three";
import { postGoRecord } from "../vscode-api";
import { encodeOverlaysToggle } from "../../schema/input-layout";
import type { OverlayFlag } from "../../messages";
import { postLog } from "../log/post";
import { useOverlayFlags } from "./overlay-flags";
import { sendRawInput, buildHomeRaw } from "./raw-input";
import {
  pillContainerStyle,
  pillBodyStyle,
  pillCaretStyle,
  inFlowPopoverStyle,
  PILL_ANCHOR_STYLE,
  groupHeadingStyle,
  DISCLOSURE_GLYPH_STYLE,
  popoverRowStyle,
  CHROME_TEXT,
  REVEALED_LIST_STYLE,
} from "./overlay-chrome";

// ---------------------------------------------------------------------------
// Shared Toggle component
// ---------------------------------------------------------------------------

type ToggleCfg = {
  flag: OverlayFlag;
  /** Initial value shown before the first buffer snapshot lands (store polarity).
   *  OMIT this for a flag whose Go-owned default is not something TS should assert —
   *  useToggleVal then falls back to `false` (render null-until-first-snapshot, same as
   *  NavGuides.tsx's `bufFlags?.x ?? false`) instead of TS authoring a stand-in value. */
  default?: boolean;
  /** Compute active (highlight) from the raw value. */
  active: (val: boolean) => boolean;
  /** The row's glyph, kept OUT of the label. It used to be part of the label string, which
   *  made its position a per-row accident (`select ⬡` trailed while the rest led) and left
   *  it in the same text run as the words — so a label that wrapped put its second line
   *  under the icon. As its own field it is always rendered first, in its own column, and
   *  the words wrap under the words. A function of the raw value for a glyph that reports
   *  state (labels' ▴/▾). */
  icon: string | ((val: boolean) => string);
  /** Label WORDS ONLY — no glyph (see `icon`). String or function of raw value. */
  label: string | ((val: boolean) => string);
  /** Title string function of active value. */
  title: (active: boolean) => string;
  /** postLog payload factory. */
  payload: (val: boolean) => Record<string, unknown>;
};

function fireToggle(cfg: ToggleCfg, val: boolean) {
  postLog("guide-btn-click", cfg.payload(val));
  postGoRecord(encodeOverlaysToggle(cfg.flag));
}

/** The value a toggle displays: the Go-owned Overlay buffer columns (the only live truth).
 *  cfg.flag keys the buffer record in store polarity. Falls back to cfg.default until the
 *  first snapshot lands. */
function useToggleVal(cfg: ToggleCfg): boolean {
  return toggleVal(useOverlayFlags(), cfg);
}

/** useToggleVal's rule as a pure function, so a caller that already holds the flags (a group
 *  header counting its members) reads them by the SAME rule instead of restating it — two
 *  copies could disagree about the pre-first-snapshot fallback and only one would be right. */
function toggleVal(bufFlags: ReturnType<typeof useOverlayFlags>, cfg: ToggleCfg): boolean {
  // ?? cfg.default only guards the (impossible) missing-key case under noUncheckedIndexedAccess;
  // every OverlayFlag is always present in the record, so `false` is preserved. When cfg omits
  // `default` (a flag whose Go default TS should not assert), this falls back to `false` until
  // the first snapshot lands — the same "null-until-first-snapshot" shape as NavGuides.tsx.
  if (bufFlags) return bufFlags[cfg.flag] ?? cfg.default ?? false;
  return cfg.default ?? false;
}

// ---------------------------------------------------------------------------
// Config table for the 10 toggle buttons
// ---------------------------------------------------------------------------

const guidelinesCfg: ToggleCfg = {
  flag: "overlays",
  default: true,
  active: (v) => v,
  icon: "▦",
  label: "overlays",
  title: (a) => (a ? "Hide overlays" : "Show overlays"),
  payload: (v) => ({ flag: "overlays", was: v }),
};

const ringsCfg: ToggleCfg = {
  flag: "tori",
  default: true,
  active: (v) => v,
  icon: "◎",
  label: "rings",
  title: (a) => (a ? "Hide polar rings" : "Show polar rings"),
  payload: (v) => ({ flag: "tori", was: v }),
};

const scenePolesCfg: ToggleCfg = {
  flag: "scenePoles",
  default: true,
  active: (v) => v,
  icon: "⊹",
  label: "scene poles",
  title: (a) => (a ? "Hide scene pole frame" : "Show scene pole frame"),
  payload: (v) => ({ flag: "scenePoles", was: v }),
};

const nodePolesCfg: ToggleCfg = {
  flag: "nodePoles",
  default: true,
  active: (v) => v,
  icon: "⊹",
  label: "node poles",
  title: (a) => (a ? "Hide node pole frames" : "Show node pole frames"),
  payload: (v) => ({ flag: "nodePoles", was: v }),
};

const selSpherePolesCfg: ToggleCfg = {
  flag: "selSpherePoles",
  default: true,
  active: (v) => v,
  // Was `select ⬡` — the one row whose glyph trailed its words. It leads now, like the rest.
  icon: "⬡",
  label: "select",
  title: (a) => (a ? "Hide select-sphere poles" : "Show select-sphere poles"),
  payload: (v) => ({ flag: "selSpherePoles", was: v }),
};

const handholdsCfg: ToggleCfg = {
  flag: "handholds",
  default: true,
  active: (v) => v !== false,
  icon: "⊙",
  label: "grips",
  title: (a) => (a ? "Hide rotation grips" : "Show rotation grips"),
  payload: (v) => ({ flag: "handholds", was: v }),
};

const globalLabelsCfg: ToggleCfg = {
  flag: "labelsGlobal",
  default: false,
  active: (v) => !v,
  icon: (v) => (v ? "▴" : "▾"),
  label: "labels",
  title: (a) => (a ? "Hide labels" : "Show labels"),
  payload: (v) => ({ flag: "labelsGlobal", wasHidden: v }),
};

// The NODE-LOCAL drawings. Everything above this line is scene furniture drawn AROUND the
// nodes; these six are the node itself, and until now nothing could turn any of them off.
// Each is `active: (v) => v` and default-on: the flag says "drawn", and a node with all six
// off simply is not drawn.

const nodeBodyCfg: ToggleCfg = {
  flag: "nodeBody",
  default: true,
  active: (v) => v,
  icon: "●",
  label: "body",
  title: (a) => (a ? "Hide node bodies" : "Show node bodies"),
  payload: (v) => ({ flag: "nodeBody", was: v }),
};

const nodeRingCfg: ToggleCfg = {
  flag: "nodeRing",
  default: true,
  active: (v) => v,
  icon: "○",
  label: "ring",
  title: (a) => (a ? "Hide node rings" : "Show node rings"),
  payload: (v) => ({ flag: "nodeRing", was: v }),
};

const ringPickCfg: ToggleCfg = {
  flag: "ringPick",
  default: true,
  active: (v) => v,
  icon: "◌",
  // The band that takes a ring click, painted so you can see where it is. Like every other
  // overlay this shows or hides a drawing and nothing else — the ring takes clicks either
  // way (that is select mode's job), so this can never quietly disable an interaction.
  label: "ring band",
  title: (a) => (a ? "Hide the ring's click band" : "Show the ring's click band"),
  payload: (v) => ({ flag: "ringPick", was: v }),
};

const selectionRingCfg: ToggleCfg = {
  flag: "selectionRing",
  default: true,
  active: (v) => v,
  icon: "◉",
  label: "selection",
  title: (a) => (a ? "Hide the selection ring" : "Show the selection ring"),
  payload: (v) => ({ flag: "selectionRing", was: v }),
};

const hoverRingCfg: ToggleCfg = {
  flag: "hoverRing",
  default: true,
  active: (v) => v,
  icon: "◍",
  label: "hover",
  title: (a) => (a ? "Hide the hover ring" : "Show the hover ring"),
  payload: (v) => ({ flag: "hoverRing", was: v }),
};

const reachSphereCfg: ToggleCfg = {
  flag: "reachSphere",
  default: true,
  active: (v) => v,
  icon: "⌾",
  label: "reach sphere",
  title: (a) => (a ? "Hide the reach sphere" : "Show the reach sphere"),
  payload: (v) => ({ flag: "reachSphere", was: v }),
};

// ---------------------------------------------------------------------------
// Grouped overlay rows for the popover
// ---------------------------------------------------------------------------

// `under` names the cfg a row NESTS beneath: that row renders indented and is disabled
// whenever its parent is off. View structure only — the gating that actually suppresses the
// drawing is Go-owned and lives in the renderer, so a disabled row here is never the only
// thing holding a child off.
type OverlayGroup = { heading: string; cfgs: ToggleCfg[]; under?: Partial<Record<string, ToggleCfg>> };

// EVERY overlay sits in a cluster — there is no loose row at the top level of the popover,
// so the list reads as "which part of the picture" rather than as one flat inventory. The
// clusters answer that question in one word each: what a NODE is made of, what marks the
// node you are touching, the scene furniture you navigate BY, the pole frames, the text.
//
// NODE and STATE are the new ones. The split between them is what changes the drawing
// permanently (a node's body and ring are there whatever you do) versus what appears
// because of where the pointer or the selection is right now.
const OVERLAY_GROUPS: OverlayGroup[] = [
  { heading: "NODE",   cfgs: [nodeBodyCfg, nodeRingCfg, ringPickCfg] }, // body, ring, click band
  { heading: "STATE",  cfgs: [selectionRingCfg, hoverRingCfg, reachSphereCfg] },
  { heading: "GUIDES", cfgs: [ringsCfg, handholdsCfg] },
  { heading: "POLES",  cfgs: [scenePolesCfg, nodePolesCfg, selSpherePolesCfg] },
  { heading: "LABELS", cfgs: [globalLabelsCfg] },
];

/** A single row inside the popover: square checkbox + label, fires the row's op on click.
 *  Styled to match the recommended mock (overlay-toggle-options.html): custom .cb checkbox
 *  that fills accent + ✓ when checked, with a subtle row-hover background. */
function OverlayRow({ cfg, disabled, indent }: { cfg: ToggleCfg; disabled?: boolean; indent?: boolean }) {
  const val = useToggleVal(cfg);
  const active = cfg.active(val);
  const [hover, setHover] = useState(false);
  const onClick = useCallback(
    (e: React.MouseEvent) => {
      e.stopPropagation();
      if (disabled) return;
      fireToggle(cfg, val);
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [val, disabled]
  );
  const labelText = typeof cfg.label === "function" ? cfg.label(val) : cfg.label;
  const iconText = typeof cfg.icon === "function" ? cfg.icon(val) : cfg.icon;
  return (
    <div
      onClick={onClick}
      onMouseEnter={() => setHover(true)}
      onMouseLeave={() => setHover(false)}
      title={cfg.title(active)}
      style={{
        ...popoverRowStyle(hover, !!disabled),
        paddingLeft: indent ? 20 : 6,
      }}
    >
      {/* The checkbox is the ONLY thing that fades when the row is inert. The label stays
          full strength — fading it was what made the open list hard to read — and the pill
          still says whether the master gate is on. Here the fade is on the one element whose
          job is to be clicked, so it reads as "this box is not taking clicks" rather than as
          the whole row receding. */}
      <span
        style={{
          width: 13,
          height: 13,
          flex: "0 0 auto",
          // Beside the first line too, for the same reason as the icon below it.
          alignSelf: "flex-start",
          opacity: disabled ? 0.45 : 1,
          borderRadius: 3,
          border: `1.5px solid ${active ? "#4ea1ff" : "#9a9aa6"}`,
          background: active ? "#4ea1ff" : "transparent",
          display: "grid",
          placeItems: "center",
          color: "#04101f",
          fontSize: 10,
          fontWeight: 900,
          lineHeight: "11px",
        }}
      >
        {active ? "✓" : ""}
      </span>
      {/* The icon LEADS, in a column of its own — never inside the label's text run. A
          glyph sharing that run is just another word, so a label that wrapped put its
          second line under the icon; here the icon is a sibling the text never flows
          beneath, and the words wrap under the words. Fixed-width so every row's words
          start at one x whatever glyph precedes them. */}
      {/* `alignSelf: flex-start` so a wrapped row keeps its glyph beside the FIRST line
          rather than floating to the middle of two. Identical to the row's centering while
          the label is one line, which is every row until one wraps. */}
      <span style={{ width: 11, flex: "0 0 auto", textAlign: "center", alignSelf: "flex-start" }}>
        {iconText}
      </span>
      {/* Wraps rather than widening: the row lives in a box that measures as nothing, so a
          label longer than the popover has to break onto the next line or it would just
          overflow the edge. `minWidth: 0` is what lets a flex item shrink below its own
          content — without it the default `min-width: auto` keeps the label at full width
          and the break never happens. */}
      <span style={{ minWidth: 0, overflowWrap: "break-word" }}>{labelText}</span>
    </div>
  );
}

/** One collapsible group in the popover: a clickable heading that expands to its rows.
 *
 *  Collapsed is the DEFAULT, and the heading carries an on/total count so collapsing never
 *  hides state — you can read "POLES 2/3" without expanding, which is the thing a plain
 *  dropdown would cost you. Open/closed is view-local `useState`, deliberately NOT a Go
 *  flag: which section a person has twirled open is not part of the model (no buffer
 *  column, nothing streamed, nothing persisted), unlike the overlay flags themselves, which
 *  stay Go-owned. Each ROW still reads its own flag from the buffer — the count here is a
 *  second reader of the same truth, never a cache of it. */
function OverlayGroupSection({ group, disabled }: { group: OverlayGroup; disabled?: boolean }) {
  const [open, setOpen] = useState(false);
  const [hover, setHover] = useState(false);
  const [countHover, setCountHover] = useState(false);
  const bufFlags = useOverlayFlags();
  const on = group.cfgs.filter((cfg) => cfg.active(toggleVal(bufFlags, cfg))).length;
  // Flip only the members that need flipping — every send is the SAME per-flag toggle record
  // a row click sends (encodeOverlaysToggle), so the group action introduces no second way to
  // set an overlay. Members already in the target state are left alone rather than toggled
  // twice. stopPropagation keeps this off the heading's expand/collapse.
  const onCountClick = useCallback(
    (e: React.MouseEvent) => {
      // The chip is the group's toggle ONLY while overlays are on. With the master off it
      // has no toggle to give, so it does NOT stop the click: it falls through to the
      // heading and expands the group like the words and the triangle beside it. Swallowing
      // it (the old `stopPropagation` before this check) made one part of the heading dead
      // to a click the rest of it answers.
      if (disabled) return;
      e.stopPropagation();
      const target = on === 0; // all off → turn everything on; otherwise turn everything off
      for (const cfg of group.cfgs) {
        const val = toggleVal(bufFlags, cfg);
        if (cfg.active(val) !== target) fireToggle(cfg, val);
      }
    },
    [group, bufFlags, on, disabled]
  );
  return (
    <div>
      <div
        onClick={(e) => { e.stopPropagation(); setOpen((o) => !o); }}
        onMouseEnter={() => setHover(true)}
        onMouseLeave={() => setHover(false)}
        title={open ? `Collapse ${group.heading}` : `Expand ${group.heading}`}
        style={groupHeadingStyle(hover)}
      >
        <span style={DISCLOSURE_GLYPH_STYLE}>{open ? "▼" : "▶"}</span>
        <span style={{ flex: "1 1 auto" }}>{group.heading}</span>
        {/* The count is also the group's toggle, and it is SYMMETRIC with no remembered
            state: any member on → turn them all off; all off → turn them all on. The
            tempting version ("off, then restore what was on") needs to remember which
            members were on — a cache of Go-owned flags in TS, or a new per-group flag in
            Go. Neither is worth it for something three row clicks already do, so this
            sends nothing but the per-flag toggle records the rows themselves send.
            Accented only when some member is on, so a collapsed group reads at a glance. */}
        <span
          onClick={onCountClick}
          onMouseEnter={() => setCountHover(true)}
          onMouseLeave={() => setCountHover(false)}
          title={disabled ? "" : on > 0 ? `Turn all ${group.heading} off` : `Turn all ${group.heading} on`}
          style={{
            // Accent when some member is on; otherwise the chrome's own text colour. It was
            // #6e6e78 — dimmer than anything else in the popover, so an all-off group's count
            // was the hardest thing to read in it, when "0/3" is exactly what you look for.
            color: on > 0 ? "#4ea1ff" : CHROME_TEXT,
            fontVariantNumeric: "tabular-nums",
            // Pointer either way: with the master off this chip is part of the heading's
            // expand target, so a default cursor here would say "nothing to click" over
            // something that does answer a click.
            cursor: "pointer",
            padding: "1px 4px",
            borderRadius: 4,
            background: !disabled && countHover ? "rgba(255,255,255,0.10)" : "transparent",
          }}
        >
          {on}/{group.cfgs.length}
        </span>
      </div>
      {open && (
        <div style={REVEALED_LIST_STYLE}>
          {group.cfgs.map((cfg) => {
            const parent = group.under?.[cfg.flag];
            // A nested row is dead while its parent is off — the same rule the master
            // `overlays` flag applies to every row, one level down. Read through toggleVal,
            // the same rule a row itself reads by, so parent and child cannot disagree about
            // the parent's value.
            const parentOff = !!parent && !parent.active(toggleVal(bufFlags, parent));
            return (
              <OverlayRow
                key={cfg.flag}
                cfg={cfg}
                disabled={disabled || parentOff}
                indent={!!parent}
              />
            );
          })}
        </div>
      )}
    </div>
  );
}

/** OVERLAYS CONTROL: split-button (body = master toggle, caret = popover) + popover checklist. */
export function OverlaysControl() {
  const [open, setOpen] = useState(false);
  const val = useToggleVal(guidelinesCfg);
  const active = guidelinesCfg.active(val);

  const onBodyClick = useCallback(
    (e: React.MouseEvent) => {
      e.stopPropagation();
      fireToggle(guidelinesCfg, val);
    },
    [val]
  );

  const onCaretClick = useCallback((e: React.MouseEvent) => {
    e.stopPropagation();
    setOpen((o) => !o);
  }, []);

  return (
    // The pill and the popover are SIBLINGS inside the shared-width anchor, so they always
    // come out the same width (PILL_ANCHOR_STYLE). Never nest the popover in the pill:
    // pillContainerStyle sets `overflow: hidden` to clip the chip's rounded corners, which
    // would clip the popover out of existence.
    <div style={PILL_ANCHOR_STYLE}>
      {/* Split button — labeled pill (body = master toggle, caret = popover). Accent fill
          when the master is on, neutral chip when off (overlay-toggle-options.html mock). */}
      <div
        style={{
          // Placed LAST in ThreeView's right-hand flex column, so it sits below every
          // panel above it however tall they are — the stacking is the column's, not a
          // number here that has to be re-derived whenever a panel above grows.
          ...pillContainerStyle(active),
        }}
      >
        {/* Body — master toggle. `flex: "1 1 auto"` so the LABEL takes the pill's slack and
            the caret stays at the far end, the same as the angles pill. Without it the
            caret sits right after the word and slides whenever the shared width changes —
            which is what made it look like the triangle was following the popover. */}
        <div
          onClick={onBodyClick}
          title={guidelinesCfg.title(active)}
          style={{ ...pillBodyStyle, flex: "1 1 auto" }}
        >
          Overlays
        </div>
        {/* Caret — popover toggle */}
        <div
          onClick={onCaretClick}
          title={open ? "Close overlay list" : "Open overlay list"}
          style={pillCaretStyle}
        >
          {/* Same disclosure triangles as the group headings (see OverlayGroupSection). */}
          {open ? "▲" : "▼"}
        </div>
      </div>

      {/* Popover — grouped checklist (.pop mock style: panel2 bg, border, shadow). IN FLOW
          under the pill, filling the anchor's width — not the old absolute popover at a
          fixed 150. In flow it is measured, so the anchor sizes to whichever is wider (the
          pill's label or a group heading) and BOTH come out at that width. The rows measure
          as nothing (REVEALED_LIST_STYLE), so expanding a group still changes neither. */}
      {open && (
        <div style={inFlowPopoverStyle()}>
          {/* No dimming for the master-off state: the PILL is that indicator — unlit chip
              means overlays are off — and a second, fainter copy of the same fact inside the
              popover only made the list hard to read. The rows stay inert (`disabled`), they
              just no longer fade to say so; a checkmark still shows each flag's real value. */}
          {OVERLAY_GROUPS.map((group) => (
            <OverlayGroupSection key={group.heading} group={group} disabled={!active} />
          ))}
        </div>
      )}
    </div>
  );
}

// ---------------------------------------------------------------------------
// Widgets: Home button
// ---------------------------------------------------------------------------

/** HOME BUTTON: reframes the camera to fit all nodes in view. */
export function HomeButton({
  cameraRef,
  aspect,
}: {
  cameraRef: React.MutableRefObject<THREE.PerspectiveCamera | null>;
  aspect: number;
}) {
  const onClick = useCallback((e: React.MouseEvent) => {
    e.stopPropagation();
    const cam = cameraRef.current;
    if (!cam) return;
    // Home is a COMMAND to Go. TS sends only render context (fov + aspect); Go frames the
    // scene from its OWN node geometry, installs the pose in the gesture FSM, and streams it
    // back via the buffer's Camera row (BufferCamera). Because the FSM's own pose becomes the
    // framed pose, the next pan/zoom/rotate builds on it (no snap-back). We do NOT mutate the
    // three.js camera or seed a computed pose here.
    sendRawInput(buildHomeRaw(cam.fov, aspect));
  }, [cameraRef, aspect]);

  return (
    <div
      onClick={onClick}
      title="Fit diagram in view"
      style={{
        // Placed by ThreeView's right-hand flex column, not by its own top/right. That
        // column stretches its widgets to one width so the pills match each other; this
        // is not a pill, so it opts out and stays as wide as "⌂ fit".
        alignSelf: "flex-end",
        background: "rgba(0,0,0,0.55)",
        borderRadius: 6,
        padding: "3px 7px",
        cursor: "pointer",
        pointerEvents: "auto",
        zIndex: 20,
        color: "#ddd",
        fontSize: 11,
        fontFamily: "monospace",
        userSelect: "none",
        display: "flex",
        alignItems: "center",
        gap: 4,
      }}
    >
      ⌂ fit
    </div>
  );
}


