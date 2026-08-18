---
branch: task/remove-subscriptions
---

# A panel is a goroutine; TS draws it on the camera rectangle

## The target

Each panel is a goroutine that owns its own state and streams what to draw. TS draws it and
nothing else: no derived rows, no panel state, no subscription.

A panel sits on the camera rectangle — the editor view — like a thing on the lens. Its
coordinates are viewport coordinates and never mention the camera, so there is nothing to
recompute when the camera moves and nothing that could move it. Not "kept fixed"; simply
not expressed in world terms.

## What each control is

One control — a checkmark, a button, a readout — is one row, and Go owns three things about
it:

- its **rect** in viewport coordinates
- its **value** (checked, the number, the text of a readout)
- its **label**

One column per thing, run-valued where the count varies, the shape the tilt arrows and edge
beads already use. The flag values are ALREADY per-flag columns
(`COL_STREAM_OVERLAY_TORI`, `..._RING_PICK`, …) — what is new is the rect.

## Input needs no new vocabulary

`RawInputMsg` already carries `X, Y`, and the gesture FSM already uses them
(`g.PrevX, g.PrevY = ev.X, ev.Y`). The click and the rects are in the SAME viewport
coordinates, so the hit test is point-in-rect, in Go, before the pointer means anything
else.

No raycaster, no pick mesh, no projection, and **no change to the TS → Go bridge surface**.
An earlier draft of this plan claimed panel input had to become a new `RawHit` kind; that
was wrong — Go already has the coordinates.

## Drawing: it looks the way it looks now

The panel is drawn in the canvas, in an orthographic overlay pass with its own camera that
never moves, so viewport pixels map 1:1 and nothing skews under perspective.

**Appearance is not a design question — it must match what is on screen today**, which the
CSS already fully specifies: `ui-sans-serif, system-ui, sans-serif` at 11–13px, weights
600/700, `#222` on `#fff`, 1px `#ddd` borders, 3–6px radii, `tabular-nums` on numeric
readouts (`src/webview/webview-toolbar.css`).

That determines the method rather than leaving it open: text is rasterised by the SAME
engine that draws it now — the browser's 2D canvas `fillText` with that font string, blitted
as a texture. A glyph atlas or SDF would not match; SDF changes edge rendering and an atlas
changes metrics and kerning. Boxes, borders and radii are rects, and a rect is a matrix,
which is what every instanced thing here already receives from Go.

Two things this makes mine to get right, not questions to ask:

- rasterise at `devicePixelRatio` and draw at CSS-pixel size, or canvas text is soft on a
  retina display — the usual reason canvas text looks worse than DOM text, and avoidable
- re-rasterise a label only when its string changes; the string arrives on its own column, so
  "changed" is that column changing — no cache and no comparison of derived objects

## What this dissolves

Nothing needs telling that a value changed. A checkmark is drawn in the frame loop from its
own column, exactly as a node or an edge is. That removes the reason the subscriptions exist
rather than replacing them, and it removes the derived caches with them —
`cachedRuleRows`, `cachedTiltVectorRows`, and the overlay/panel bundle objects that
reassemble columns Go already streams separately.

## Order

1. **SpeedSlider** — one value, one control, no rows. Establishes the rect column, the
   orthographic pass, the text rasterisation and the point-in-rect hit test end to end, on
   the smallest surface that exercises all of them.
2. Stop and judge. Five panels is a lot of surface to commit to an unproven shape.
3. **TiltVectorButtons** — rows that vary with node count; the first run-valued panel columns.
4. **NodeRulesPanel** — the largest, and the one deriving rows from nodes AND edges today.
5. **AngleDropdown**, **NodesDropdown**, **OverlaysDropdown** — 8 of the 22 hook sites and
   both bundle objects live in these three.

## Verification

- **Headless**: drive the real binary and read the panel's columns off the wire — its rects,
  values and labels — the same way the bead and tilt columns were checked. A silent column
  and a clean build look identical to the guards, so read the wire.
- **In the editor**: the panel looks the same as it does today; it does not move under orbit
  or zoom; a click toggles the right control; and a breadcrumb on the panel goroutine names
  the control row it hit.

## Risks

- **Coordinate space.** Go hit-tests rects TS draws. CSS pixels vs device pixels, or a
  disagreement about the canvas origin, shows up as clicks landing on the wrong control and
  reads like a logic bug. The click coordinates Go already receives fix the convention — the
  rects must use the same one.
- **A control mid-gesture** — a slider being dragged — is state the panel goroutine owns, and
  the round trip now crosses the bridge. A slider that lags its own drag is the first thing
  anyone will notice.
- **Text is the only genuinely new machinery.** Everything else reuses what already works.
