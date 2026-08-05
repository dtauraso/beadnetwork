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
  /** Label string or function of raw value. */
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
  label: "▦ overlays",
  title: (a) => (a ? "Hide overlays" : "Show overlays"),
  payload: (v) => ({ flag: "overlays", was: v }),
};

const ringsCfg: ToggleCfg = {
  flag: "tori",
  default: true,
  active: (v) => v,
  label: "◎ rings",
  title: (a) => (a ? "Hide polar rings" : "Show polar rings"),
  payload: (v) => ({ flag: "tori", was: v }),
};

const scenePolesCfg: ToggleCfg = {
  flag: "scenePoles",
  default: true,
  active: (v) => v,
  label: "⊹ scene poles",
  title: (a) => (a ? "Hide scene pole frame" : "Show scene pole frame"),
  payload: (v) => ({ flag: "scenePoles", was: v }),
};

const nodePolesCfg: ToggleCfg = {
  flag: "nodePoles",
  default: true,
  active: (v) => v,
  label: "⊹ node poles",
  title: (a) => (a ? "Hide node pole frames" : "Show node pole frames"),
  payload: (v) => ({ flag: "nodePoles", was: v }),
};

const selSpherePolesCfg: ToggleCfg = {
  flag: "selSpherePoles",
  default: true,
  active: (v) => v,
  label: "select ⬡",
  title: (a) => (a ? "Hide select-sphere poles" : "Show select-sphere poles"),
  payload: (v) => ({ flag: "selSpherePoles", was: v }),
};

const handholdsCfg: ToggleCfg = {
  flag: "handholds",
  default: true,
  active: (v) => v !== false,
  label: "⊙ grips",
  title: (a) => (a ? "Hide rotation grips" : "Show rotation grips"),
  payload: (v) => ({ flag: "handholds", was: v }),
};

const globalLabelsCfg: ToggleCfg = {
  flag: "labelsGlobal",
  default: false,
  active: (v) => !v,
  label: (v) => `${v ? "▴" : "▾"} labels`,
  title: (a) => (a ? "Hide labels" : "Show labels"),
  payload: (v) => ({ flag: "labelsGlobal", wasHidden: v }),
};

// cascadeLinksCfg has no `default` — its Go-owned default (off) is not asserted here; see
// useToggleVal's fallback and ToggleCfg.default's doc above.
const cascadeLinksCfg: ToggleCfg = {
  flag: "cascadeLinks",
  active: (v) => v,
  label: "⇉ cascade links",
  title: (a) => (a ? "Hide cascade-link overlay" : "Show cascade-link overlay"),
  payload: (v) => ({ flag: "cascadeLinks", was: v }),
};

// ---------------------------------------------------------------------------
// Grouped overlay rows for the popover
// ---------------------------------------------------------------------------

// `under` names the cfg a row NESTS beneath: that row renders indented and is disabled
// whenever its parent is off. View structure only — the gating that actually suppresses the
// drawing is Go-owned and lives in the renderer, so a disabled row here is never the only
// thing holding a child off.
type OverlayGroup = { heading: string; cfgs: ToggleCfg[]; under?: Partial<Record<string, ToggleCfg>> };

const OVERLAY_GROUPS: OverlayGroup[] = [
  { heading: "GUIDES", cfgs: [ringsCfg, handholdsCfg] },
  { heading: "POLES",  cfgs: [scenePolesCfg, nodePolesCfg, selSpherePolesCfg] },
  { heading: "LABELS", cfgs: [globalLabelsCfg] },
  { heading: "EDGES",  cfgs: [cascadeLinksCfg] },
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
  return (
    <div
      onClick={onClick}
      onMouseEnter={() => setHover(true)}
      onMouseLeave={() => setHover(false)}
      title={cfg.title(active)}
      style={{
        display: "flex",
        alignItems: "center",
        gap: 7,
        padding: "4px 6px",
        paddingLeft: indent ? 20 : 6,
        opacity: disabled ? 0.45 : 1,
        cursor: disabled ? "default" : "pointer",
        color: "#e7e7ea",
        borderRadius: 5,
        background: !disabled && hover ? "rgba(255,255,255,0.05)" : "transparent",
        userSelect: "none",
        fontSize: 11.5,
      }}
    >
      <span
        style={{
          width: 13,
          height: 13,
          flex: "0 0 auto",
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
      <span>{labelText}</span>
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
      e.stopPropagation();
      if (disabled) return;
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
        style={{
          display: "flex",
          alignItems: "center",
          gap: 5,
          fontSize: 9.5,
          textTransform: "uppercase",
          letterSpacing: "0.05em",
          color: "#9a9aa6",
          padding: "5px 6px 4px",
          cursor: "pointer",
          borderRadius: 5,
          background: hover ? "rgba(255,255,255,0.05)" : "transparent",
        }}
      >
        {/* ▶/▼ (U+25B6/U+25BC), not the ▸/▾ small variants: those render as thin arrowheads
            in several of the fonts this stack falls back to, which reads as a link chevron
            rather than a disclosure triangle. */}
        <span style={{ fontSize: 8, width: 8, flex: "0 0 auto" }}>{open ? "▼" : "▶"}</span>
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
            color: on > 0 ? "#4ea1ff" : "#6e6e78",
            fontVariantNumeric: "tabular-nums",
            cursor: disabled ? "default" : "pointer",
            padding: "1px 4px",
            borderRadius: 4,
            background: !disabled && countHover ? "rgba(255,255,255,0.10)" : "transparent",
          }}
        >
          {on}/{group.cfgs.length}
        </span>
      </div>
      {open && group.cfgs.map((cfg) => {
        const parent = group.under?.[cfg.flag];
        // A nested row is dead while its parent is off — the same rule the master `overlays`
        // flag applies to every row, one level down. Read through toggleVal, the same rule a
        // row itself reads by, so parent and child cannot disagree about the parent's value.
        const parentOff = !!parent && !parent.active(toggleVal(bufFlags, parent));
        return (
          <OverlayRow key={cfg.flag} cfg={cfg} disabled={disabled || parentOff} indent={!!parent} />
        );
      })}
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

  const fontStack = '-apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif';

  return (
    <>
      {/* Split button — labeled pill (body = master toggle, caret = popover). Accent fill
          when the master is on, neutral chip when off (overlay-toggle-options.html mock). */}
      <div
        style={{
          position: "absolute",
          // Stacked below the distance-home panel (DistanceHomePanel, absolute top:66,
          // ~55px tall) which itself sits below the fit/home button (top:44). All three
          // share this same containing block (ThreeView's containerRef, absolute inset:0)
          // and the same `position: absolute` scheme, so they anchor/scroll together.
          top: 128,
          right: 12,
          zIndex: 20,
          pointerEvents: "auto",
          display: "flex",
          alignItems: "stretch",
          borderRadius: 6,
          overflow: "hidden",
          fontSize: 11,
          fontWeight: 600,
          fontFamily: fontStack,
          background: active ? "#4ea1ff" : "#34343d",
          border: `1px solid ${active ? "#4ea1ff" : "#3a3a44"}`,
          color: active ? "#04101f" : "#9a9aa6",
          userSelect: "none",
        }}
      >
        {/* Body — master toggle */}
        <div
          onClick={onBodyClick}
          title={guidelinesCfg.title(active)}
          style={{ padding: "3px 9px", cursor: "pointer", display: "flex", alignItems: "center" }}
        >
          Overlays
        </div>
        {/* Caret — popover toggle */}
        <div
          onClick={onCaretClick}
          title={open ? "Close overlay list" : "Open overlay list"}
          style={{
            padding: "3px 7px 3px 4px",
            cursor: "pointer",
            display: "flex",
            alignItems: "center",
            fontSize: 9,
            opacity: 0.85,
          }}
        >
          {/* Same disclosure triangles as the group headings (see OverlayGroupSection). */}
          {open ? "▲" : "▼"}
        </div>
      </div>

      {/* Popover — grouped checklist (.pop mock style: panel2 bg, border, shadow). */}
      {open && (
        <div
          style={{
            position: "absolute",
            top: 156,
            right: 12,
            zIndex: 21,
            pointerEvents: "auto",
            width: 150,
            background: "#2f2f37",
            border: "1px solid #3a3a44",
            borderRadius: 8,
            padding: 6,
            boxShadow: "0 8px 24px rgba(0,0,0,0.5)",
            fontFamily: fontStack,
            userSelect: "none",
          }}
        >
          <div style={{ opacity: active ? 1 : 0.4, transition: "opacity 0.12s ease" }}>
            {OVERLAY_GROUPS.map((group) => (
              <OverlayGroupSection key={group.heading} group={group} disabled={!active} />
            ))}
          </div>
        </div>
      )}
    </>
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
        position: "absolute",
        top: 44,
        right: 12,
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


