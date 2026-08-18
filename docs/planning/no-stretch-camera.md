---
branch: task/no-stretch-camera
---

# Resizing the window reveals more diagram; it never rescales it

## The rule

Dragging the editor's panel divider changes how much of the diagram you can see. It does not
change how big the diagram is. A node keeps its size on screen; a taller editor shows more
scene above and below it, the way a map does.

The panels already obey this — they are laid out in viewport pixels and drawn on the camera
rectangle. The scene does not, and that inconsistency is the whole of this change.

## Why it stretches today

The scene camera has a fixed vertical field of view (50°, set where the `<Canvas>` is
created). Fixed vertical FOV means the world span visible top-to-bottom is constant, so a node
occupies a constant FRACTION of the viewport's height — and therefore a pixel size that scales
with that height. Widening behaves differently: the aspect grows, so more world is visible
left and right at the same scale.

So height rescales and width reveals. The rule above says both should reveal.

This is not a regression. It is the convention the camera was written with, and it predates
the panel work; what changed is that the panels stopped doing it, which made the scene's
behaviour visible by contrast.

## The single change

**The field of view is derived from the viewport, and Go owns it.**

Holding world-units-per-pixel constant while the viewport height changes means the FOV must
open as the rectangle grows:

    fov = 2 · atan(k · H / (2 · d))

where `H` is the rectangle's height in pixels, `d` the camera's distance to the pivot, and `k`
the world units per pixel the scene is drawn at.

Go already knows all three. It is told the rectangle (`scene`/`viewport`, added so the
right-anchored pills could be laid out), it owns the viewpoint and therefore `d`, and `k` is
whatever this change decides zoom means. It streams the resulting FOV on a column and the
webview applies it to the three camera, the same way it applies every other camera value.

## This also closes a local≠global

FOV is a webview-owned camera property today. It is set on the `<Canvas>`, read back off the
three camera on every pointer event, and SENT TO GO as `ev.Fov` — where Go uses it for
ray-through-NDC math and for the fit pose. So the camera's pose is Go's while one of its
parameters lives in the webview and is reported back on every event.

That is the same shape as the overlay flags and the tilt rows that were just removed: a value
Go needs, held on the other side of the bridge and echoed across it. When Go derives the FOV,
`ev.Fov` stops being an input and becomes a value Go already has.

## Open — decide before building

1. **What does zoom mean?** Zoom currently changes the camera's distance `d`. Under a derived
   FOV, either `k` stays constant and zoom remains distance-only, or `k` IS the zoom and `d`
   stays fixed. These feel different to use and the choice is not implied by the rule.
2. **What does fit do?** Fitting the diagram inherently depends on the rectangle's shape — a
   wide short editor frames a wide diagram differently from a tall one. Fit stays
   viewport-dependent under this rule; the question is only which dimension binds when the
   diagram's aspect and the rectangle's disagree.

## Ripple

- `fov` leaves the `<Canvas>` camera props and the `RawInputEvent`; `ev.Fov` loses its
  meaning as an input and its uses in `gesture_handlers.go` and `HomeFitPose` read Go's own
  value.
- A Camera-block column carries the FOV to the webview, which sets it on the three camera and
  calls `updateProjectionMatrix`.
- `DropPointFromNDC` and `DragPlaneHit` already take a fov argument; they take Go's.
- The input-layout fingerprint changes if `fov` is dropped from the raw-input record.

## Verification

- Drag the panel divider across its full range: a node's pixel size does not change, and more
  or less of the scene is visible above and below it.
- Read the FOV column off the wire at two rectangle heights and confirm it opens with height.
- Orbit, zoom and fit still behave; the fit button still frames the whole diagram.
