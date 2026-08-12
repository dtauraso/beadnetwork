---
name: enumerate-what-rides-a-channel
description: Before gating or removing a channel because of its volume, enumerate everything else that rides it. The .probe firehose and the documented Go debugging breadcrumbs shared the same files, and the cheap gate silently took both.
type: feedback
---

**Rule:** volume is a property of one PRODUCER, but a gate applies to the CHANNEL. Before turning a channel off, list every producer on it.

**The near-miss (2026-07-28):** `.probe/go-edge.jsonl` reached 1.1 GB, so the trace logs were gated behind `wirefold.probe.trace` (default off). The first version of that gate silently broke the documented Go debugging channel: breadcrumbs are not a separate file — since the binary-buffer move they ride the per-owner streams as `kind=="breadcrumb"`, and `probe-merge.sh --debug` greps them out of the exact four files the setting turned off.

The failure mode is what makes it worth remembering: a breadcrumb would fire and produce nothing, with **no error**. The natural conclusion is "my breadcrumb is broken", not "the gate ate it". Fixed by making breadcrumb rows unconditional and gating only the non-breadcrumb bulk.

**The decisive question was "who consumes this?", not "how big is it?"** Grepping `webview/` for `edge-bead` returned zero hits — the renderer reads the frame's Bead block instead. The event duplicated x/y/z/value at 67 bytes against the Bead block's 16, and its sole path was `appendFileSync` to a log. This did NOT violate "Go only emits a frame when something changes"; a bead in flight really does change every frame. What was unjustified was the second, larger copy that only a log file read.

**Corollary — gate PRODUCTION on a consumer existing, never fix it by draining.** `PacedWire.pending` grew forever whenever no per-edge stream was wired: appends were gated on the placement's GEOMETRY (`bp.Node != ""`), which says nothing about whether anyone would read the events, while the only drain sat behind an early return for a nil `streamOut`. Two independent conditions, so appends continued while nothing drained — and `newEdgeMover` never set those fields, so it covered every headless run. Fixed with `StreamsActive`, set once at wiring time. Draining unconditionally would have kept building an event per arrival and discarding it, and would have left "who consumes this?" hidden, which is what caused the bug.

See [[go-vs-coordinator-bias]].
