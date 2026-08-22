---
paths:
  - "src/Node/**/*.go"
  - "src/**/trace_event.go"
  - "src/**/trace_log.go"
  - "scripts/probe-merge.sh"
---

# Debugging the Go layer (probe breadcrumbs)

Go-side runtime debugging goes through the **DEBUG BREADCRUMB channel** — the Go analogue
of the webview's `postLog`. Call `tr.Breadcrumb(label, node, port, value string)` at a debug
site: it is a structured `Kind==KindBreadcrumb` row (`src/Node/nodeactor/owners/trace_event.go`) that the
EMITTING goroutine writes ITSELF, as a fixed-width binary record appended to the file
belonging to the item the event is about:

    topology/view/nodes/<row>/trace.bin           the node
    topology/view/nodes/<row>/interior-trace.bin  its interior
    topology/view/nodes/<row>/beads-trace.bin     its beads
    topology/view/edges/<row>/trace.bin           the edge
    topology/view/trace.bin                       the VIEW owner

One writer per file — the goroutine that owns that item — so the append needs no lock,
the same rule the block files follow. **These events do NOT cross the Go→TS seam.** They
used to ride an events section on every stream frame, which the ext host decoded and
re-encoded as text; that whole path is gone, along with `buffer-log.ts` and the TS event
decoders. A stream frame now carries only a tick and the layout fingerprint.

Read them with `scripts/probe-merge.sh --debug`, which decodes every `trace.bin` at READ
time via `src/Trace/readtrace` and filters to `kind=="breadcrumb" && debug==true` —
separate from genuine stderr errors (`.probe/go-errors.log`, still plain text). To read one
item on its own: `go run ./src/Trace/readtrace topology/view/nodes/3/trace.bin`.

Do NOT scatter `fmt.Fprintf(os.Stderr, ...)` for diagnosis; use a breadcrumb. Keep it
SPARSE — it is a debug tool for control events, not a per-tick firehose (see the log-flood
lesson). It is a cheap no-op when no scene root is set (headless tests).

## What the trace setting gates

The `wirefold.probe.trace` setting (default **off**) gates the non-breadcrumb bulk of the
trace files plus `ts.log` — the per-tick firehose (recv/send/edge-bead/node-geometry/etc.)
that once grew `go-edge.log` past a gigabyte. It does NOT gate breadcrumb rows: Go reads
`WIREFOLD_PROBE_TRACE` once at startup into `trace.TraceEnabled()`, and `Log.Append` skips
every non-breadcrumb event when it is off, so `scripts/probe-merge.sh --debug` and the
always-on error logs (`go-errors.log`/`ts-errors.log`/`handler-error-last.log`) work out of
the box on a fresh install with no setting change. Turn the setting on only when you need
the full per-tick trace, not just breadcrumbs.

## Source-gated edge-bead trace

The highest-volume of these — `KindEdgeBead`, emitted per in-flight bead per tick by
`src/Node/BeadAnimation/bead_line_drive.go`'s `stepAll` — is gated at the SOURCE, not just at the TS
write. `stepAll` reads a package-level `edgeBeadTraceEnabled` bool set ONCE at process
startup from the `WIREFOLD_EDGE_BEAD_TRACE` env var (read once before any goroutine starts, the same shape `trace.TraceEnabled` uses); the
ext host (`src/extension/runCommand.ts`) sets it from the SAME
`isProbeTraceEnabled()` that gates the TS-side write, so there is one source of truth for
the setting. With tracing off, Go never builds the event at all, so nothing reaches the item trace file. `KindBreadcrumb` and
`KindArrive` are NOT gated by this flag and always emit; `LiveBeadRow`/the bead
block file that actually renders beads reads neither flag and is unaffected.

On editor hang/decouple/compound symptoms, read the `.probe` ERROR logs first
(`memory/feedback/debugging/probe-logs/feedback_runner_errors_probe_first.md`). For intermittent UI bugs, add cheap
runtime breadcrumbs + a repro before theorizing
(`memory/feedback/debugging/investigation-method/feedback_runtime_breadcrumbs_beat_static_analysis.md`).
