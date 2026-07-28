---
branch: task/edge-stream-volume
---

# Why does the edge stream emit 734 KB/s?

## Measured

Taken live on 2026-07-28, one editor session running the shipped 9-node topology,
no interaction — idle playback only:

| Log | Size after one long session |
|---|---|
| `.probe/go-edge.jsonl` | **1.1 GB** |
| `.probe/go-interior.jsonl` | 34 MB |
| `.probe/ts.jsonl` | 1.8 MB |
| `.probe/go-node.jsonl` | 900 KB |
| `.probe/go.jsonl` (view) | 14 KB |

Rate, measured over 10s with `stat -f%z`: **580914 -> 1332944 bytes = 734 KB/s,
~258 MB/hour.** The edge stream is ~30x everything else combined.

Every line is the same shape — a per-bead position sample:

```
{"ts_ms":1785224605871,"src":"go","kind":"edge-bead","node":"","port":"","value":0,
 "x":979.02,"y":-369.65,"z":-505.05,"f":0.887...}
```

## The question

Is this rate INHERENT (N beads x M frames/sec is genuinely what an edge stream carries,
and the log is simply recording every frame), or is something emitting per-frame where
it should emit per-change?

Worth checking, in order:

1. Who emits `kind:"edge-bead"` — the Go edge mover's own trace call, and at what
   cadence relative to its clock cycle.
2. Whether the ext host writes EVERY decoded event to `.probe` or is supposed to
   sample/filter. `runCommand.ts`'s `appendFileSync` sites are the write path.
3. Whether `node`/`port`/`value` being empty/zero on every line means the row is
   carrying fields it does not need (cheap win on volume, not on rate).
4. Whether the VIEW stream's 14 KB vs the edge stream's 1.1 GB reflects a real
   difference in what changes, or a missing "only emit when something changed" rule that
   the view path has and the edge path does not. CLAUDE.md says Go "only ever emits a
   frame when something changes, and that stays true" — a bead in flight DOES change
   every frame, so this may be legitimate; confirm rather than assume.

## Constraint

CLAUDE.md's breadcrumb guidance already names the log-flood lesson: keep the debug
channel SPARSE, "a debug tool for control events, not a per-tick firehose". The edge
stream is not a breadcrumb channel, but the `.probe` decode of it is subject to the same
rule.

## Related

`task/probe-log-rotation` (merged separately) bounds the SYMPTOM — logs now rotate per
run and VS Code no longer watches them. It deliberately does not touch the emission
rate. If the rate is legitimate, rotation is the whole fix and this branch closes with
that finding recorded.
