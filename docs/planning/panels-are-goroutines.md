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

## A widget is not the thing it wraps

`SpeedSlider` looks like a continuous slider and is not one: `min=0`, `max=5`, `step=1`. It
picks an INDEX into six fixed settings — `0`, ¼, ½, ¾, `1`, `2`. Six positions, one current.
The six are already drawn by us directly beneath it, as labelled ticks with the current one
bold; the `<input type="range">` above them is a widget wrapped around a six-way choice and
contributes nothing to the model.

The same holds for the rules panel's `<input type="checkbox">`: a checkbox is a rect, a mark,
and a boolean Go owns.

So there is no native appearance to reproduce, because the native control is not the thing.
Reading `type="range"` as "this panel's look is platform-drawn and cannot be matched" was the
same error this whole change exists to remove — mistaking the implementation that grew around
a thing for the thing.

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

**Appearance is not a design question — the canvas version must look the same as the one on
screen today.** Each panel is compared against the current one before its commit lands.

For everything our CSS specifies that is exact, because the same rasteriser draws it. For the
chrome a native widget contributed — a range track and thumb, a checkbox mark — we draw the
equivalent and match what is there now as closely as the platform allows; that chrome then
belongs to us, like every other rect in the scene.

The CSS specifies: `ui-sans-serif, system-ui, sans-serif` at 11–13px, weights
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

## Order — one panel per commit, all of them, then merge

1. **SpeedSlider** — six positions, one current. Small, but it exercises everything: the rect
   column, run-valued rows, the orthographic pass, text rasterisation (including the stacked
   fractions) and point-in-rect hit-testing.
2. **TiltVectorButtons** — two buttons and a readout; rows that vary with node count.
3. **AngleDropdown** — rows from the same tilt columns; no native controls.
4. **NodesDropdown** — four separate flag reads today.
5. **OverlaysDropdown** — 15 overlay flags and 11 panel flags, both bundle objects.
6. **FitButton** — no subscription, so "remove the subscriptions" never named it, but it is a
   DOM pill in the same right-hand column as the three dropdowns. Leaving it there keeps that
   column half DOM and keeps the offset scaffolding alive; it is a rect and a label, and it
   already speaks raw-input.
7. **Tabs** — the scene tab strip. Not a panel on the stack and easy to miss for that reason,
   but `Tabs/tab-state.ts` is one of the subscriptions, with its own cached `TabsState`; a tab
   is a rect, a name and a selected flag, the same as everything else here.
8. **NodeRulesPanel** — the largest: rows derived from nodes AND edges, six checkboxes,
   eleven buttons.
9. **BufferLabelOverlay** — the node-name pills over the scene. Not a panel and not a
   subscription, so nothing named it, but it is still DOM: one absolutely-positioned `<div>`
   per node, inline-styled, kept in place by projecting each label anchor to screen pixels
   every frame. It is text at a rect, which is what the canvas draws now — and it is the last
   thing making the projector push positions out to React at all.
10. The subscriptions and derived caches are then unreferenced; delete them and empty
   `check-no-webview-state.sh`'s allowlist, per `remove-subscriptions.md`.

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
