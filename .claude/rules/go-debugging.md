---
paths:
  - "src/Node/**/*.go"
  - "src/Buffer/**/*.go"
  - "scripts/probe-merge.sh"
---

# Debugging the Go layer (probe breadcrumbs)

Go-side runtime debugging goes through the **DEBUG BREADCRUMB channel** — the Go analogue
of the webview's `postLog`. Call `tr.Breadcrumb(label, node, port, value string)` at a debug
site: it is a structured `Kind==KindBreadcrumb` row on the EMITTING goroutine's own
per-owner content-buffer stream (node/edge/interior/VIEW — no per-node stream emits
onto the VIEW stream instead), decoded by the ext host exactly like every other
buffer-carried trace event (`buffer-log.ts`'s `"breadcrumb"` case) — there is no
separate JSON-on-stdout debug sink. Read it with `scripts/probe-merge.sh --debug`, which
filters the buffer-decoded `.probe` logs (`go.log`/`go-node.log`/`go-edge.log`/
`go-interior.log`) to `kind=="breadcrumb" && debug==true` — separate from genuine
stderr errors (`go-errors.log`). Do NOT scatter `fmt.Fprintf(os.Stderr, ...)` for
diagnosis; use a breadcrumb. Keep it SPARSE — it is a debug tool for control events, not
a per-tick firehose (see the log-flood lesson). It is a cheap no-op when no stream is
wired (headless tests).

## What the trace setting gates

The `wirefold.probe.trace` setting (default **off**) gates the non-breadcrumb bulk of
these same four files plus `ts.log` — the per-tick firehose (recv/send/edge-bead/
node-geometry/etc.) that once grew `go-edge.log` past a gigabyte. It does NOT gate
breadcrumb rows: every write site decodes the full frame regardless and appends
breadcrumb-only lines when the setting is off, so `scripts/probe-merge.sh --debug` and the
always-on error logs (`go-errors.log`/`ts-errors.log`/`handler-error-last.log`) work
out of the box on a fresh install with no setting change. Turn the setting on only when
you need the full per-tick trace, not just breadcrumbs.

## Source-gated edge-bead trace

The highest-volume of these — `KindEdgeBead`, emitted per in-flight bead per tick by
`src/Node/BeadAnimation/bead_line_drive.go`'s `stepAll` — is gated at the SOURCE, not just at the TS
write. `stepAll` reads a package-level `edgeBeadTraceEnabled` bool set ONCE at process
startup from the `WIREFOLD_EDGE_BEAD_TRACE` env var (same "one env var, read once before
any goroutine starts" shape as `WIREFOLD_STREAM_FDS` — see `src/runtopology/streamwire/stream_fds.go`); the
ext host (`src/extension/runCommand.ts`) sets it from the SAME
`isProbeTraceEnabled()` that gates the TS-side write, so there is one source of truth for
the setting. With tracing off, Go never appends the event to the frame at all — TS
previously decoded and discarded it every tick regardless. `KindBreadcrumb` and
`KindArrive` are NOT gated by this flag and always emit; `LiveBeadRow`/the Bead-block
buffer path that actually renders beads reads neither flag and is unaffected.

On editor hang/decouple/compound symptoms, read the `.probe` ERROR logs first
(`memory/feedback/debugging/probe-logs/feedback_runner_errors_probe_first.md`). For intermittent UI bugs, add cheap
runtime breadcrumbs + a repro before theorizing
(`memory/feedback/debugging/investigation-method/feedback_runtime_breadcrumbs_beat_static_analysis.md`).
