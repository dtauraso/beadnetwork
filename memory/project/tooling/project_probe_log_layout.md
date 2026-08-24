---
name: project-probe-log-layout
description: Runtime logs land in five .probe/ JSONL files (go/go-errors/go-debug/ts/ts-errors) with a shared ts_ms+src+step envelope; probe-merge.sh derives unified views
metadata:
  type: project
---

Editor/runtime diagnostics are written to five files under `.probe/` — `go.jsonl` (buffer-decoded trace events, src:"buf"), `go-errors.jsonl` (Go failures from stderr, src:"go"), `go-debug.jsonl` (Go DEBUG BREADCRUMB channel, src:"go-debug"), `ts.jsonl` (webview+ext logs, src:"ts-webview"/"ts-ext"), `ts-errors.jsonl` (window/unhandled/render errors).

**The breadcrumb RULE — where to call `tr.Breadcrumb`, why not `fmt.Fprintf(os.Stderr, ...)`, and keep-it-SPARSE — lives in `.claude/rules/go-debugging.md`**, a path-scoped rule that loads on demand when Go files under `nodes/` or `Buffer/` are read (not at every session start). It is deliberately not restated here: this file previously carried a second copy of the routing mechanics, so a change to `tryParseBreadcrumb` or the sink wiring needed two edits and would silently half-rot. This file covers only what CLAUDE.md does not: the file layout, the envelope, and the freshness trap.

**Envelope.** Every line carries `{ts_ms, src, step?}` — `ts_ms` is `Date.now()` wall-clock (cross-process comparable on one machine), `step` is the Go event ordinal, present only on Go-derived lines. Go's `marshalEvent`/canonical form is untouched (contract fixture `trace-events.jsonl` pins it); the envelope is added extension-side at the disk-write boundary.

**Reading across files.** `scripts/probe-merge.sh` (no-arg = all by `ts_ms`; `--errors`, `--step N`, `--go`, `--ts`, `--debug`). Retired filenames: `phase4-pump.jsonl`→`go.jsonl`, `webview-log.jsonl`→split.

**Gating (`beadnetwork.probe.trace`, default off).** The bulk, per-tick trace rows in
`go.jsonl`/`go-node.jsonl`/`go-edge.jsonl`/`go-interior.jsonl`/`ts.jsonl` only get
appended when this VS Code setting is on (they once grew `go-edge.jsonl` past a
gigabyte). Breadcrumb rows (`kind=="breadcrumb"`) are the exception — always appended to
those same four Go files regardless of the setting, since CLAUDE.md's debug channel must
work out of the box. The `-errors.jsonl` files and `handler-error-last.json` are also
always written. `--debug` and `--errors` are therefore unaffected by the setting;
`--go`/`--ts` read near-empty files until it's turned on.

The `edge-bead` source gate is GONE. `KindEdgeBead`, the package-level
`edgeBeadTraceEnabled` bool, and the `WIREFOLD_EDGE_BEAD_TRACE` env var the ext host set
from `isProbeTraceEnabled()` at spawn were all deleted in `033ef6f99`, along with the 19
other event kinds that duplicated a column — the kind had nothing left to gate. Five kinds
remain: recv, fire, send, arrive, breadcrumb. That commit removed the only line that set
any trace env var on the Go child, so `beadnetwork.probe.trace` no longer reaches Go at
all: it gates `ts.log`, while Go's own `traceEnabled` reads `BEADNETWORK_PROBE_TRACE` from
the environment it was spawned in and nothing sets it.

**Freshness caveat (the trap).** These files are written by the LIVE editor run and can be minutes stale — check the last `ts_ms` against now before concluding anything. Several diagnoses were derailed by reading a stale log that did not contain the live failure.
