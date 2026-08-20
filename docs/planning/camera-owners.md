# The camera's seam becomes owners, not one pose

## The target

The camera crosses as three INDEPENDENT owners, each writing its own primitive
files on its own cadence, and the renderer re-reads only the owner that moved.
No pose snapshot, because there is nothing to tear between.

## What the code says the owners are

Measured against the gesture handlers, not assumed:

| gesture | writes | leaves alone |
|---|---|---|
| wheel zoom (`PanViewpoint`) | `Pivot` | R, Pos, Up |
| pan (`PanViewpoint`) | `Pivot` | R, Pos, Up |
| `ZoomViewpoint` | `R` | Pivot, Pos, Up |
| orbit (`Rotate`) | `Pos`, `Up` together | Pivot, R |

`Rotate` is the only thing that couples two of them, which is why `Pos`+`Up` are
ONE owner rather than two. Wheel zoom only READS `R`, to floor its step.

## Why this matters, measured

The renderer reads all eight values every frame. During a wheel zoom, three of
them change. The other five are re-read at 60fps to observe that they did not
move.

The cost of a read on this hop, from `.probe/ts.log` while zooming:

    avg 5.3–7.0 ms per batch of 8      worst 26–41 ms

That is ~0.8ms per fetch against 9.3µs for a raw file read — the webview's
resource protocol is ~80x the filesystem. The average fits in a 16.7ms frame;
the worst is two to three frames, and that variance is what reads as jerk.

Eight concurrent fetches cost exactly eight times one, so they are not
overlapping. Whether that is fixable is unknown and worth one measurement
before any restructuring: if eight can be made to overlap, a batch costs ~0.8ms
and there is nothing left to fix.

## What the split buys

Per-owner generation files let the renderer skip an owner that has not moved: a
wheel zoom becomes one owner's reads plus two generation checks, instead of
eight full reads.

It also dissolves the snapshot question. Files were said to lose the frame
boundary, so a batch could mix two instants. Between independent owners there
is nothing to mix — no gesture writes `R` and `Pivot` together, so reading one
newer than the other is not inconsistent. Consistency is required only WITHIN
an owner, which is where `Pos` and `Up` already sit. The frame was buying a
shared instant across values that never change together.

## Also outstanding, and separate

The renderer derives the camera from polar inputs: two `anglesToWorldOffset`
calls and a fov from focal length, all of which Go already has as
`Camera.AnglesToWorldOffset` and `Camera.FovDegForHeight`. That is the render
tree computing geometry, which MODEL.md forbids, and it predates the file
transport — the buffer columns were polar too.

Go should write what the renderer ASSIGNS: position x/y/z, up x/y/z, fov. Then
TS holds no trigonometry and no `FOCAL_PIXELS`.

This does NOT fix the jerk — seven reads instead of eight is not a measurable
difference. It is worth doing on its own terms and should be judged on those.

## Verification

Drive the editor and zoom. The per-batch timing goes back in temporarily
(`postLog("camera.read-batch", …)`, read with `.probe/ts.log`) and comes out
once the number is known. `bash scripts/stop-checks.sh` empty.
