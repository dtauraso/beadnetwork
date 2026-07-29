# Audit baseline — settled findings, do not re-report

This file is the permanent record of findings from `audit-blast-radius`, `audit-priors-fit`,
and `audit-grep-load` that have already been judged **deliberate and structural** — the
intrinsic cost of this codebase's architecture, not a defect. Every audit subagent must read
this file FIRST and must NOT re-report anything listed here. Report only NEW deviations from
this baseline (a fingerprint going out of sync, a clone family diverging in behavior instead
of just being clone boilerplate, a var map gaining a post-init write, etc.).

Each entry below was verified against the code before being recorded here (grep/read, not
prose-trusted). If code changes such that an entry no longer matches, update or remove the
entry — don't leave it stale.

## 1. Node-kind clone families are deliberate

- `nodes/Time/` vs `nodes/TimeStart/` — `node.go` is 172 vs 174 lines, ~30 lines of diff
  (mostly comments/naming). `nodes/TimeStart/node.go:10` states outright: "TimeStart is a
  clone of the Time kind (see nodes/Time) — identical behavior ... split out so node 2's
  future divergence has a home of its own." This is intentional forking-in-place, not
  copy-paste drift.
- `nodes/pulse/` vs `nodes/PulseLeft/` vs `nodes/PulseRight/` — `node.go` is 138/139/139
  lines, ~30 lines of pairwise diff. Same clone-for-future-divergence pattern.
- `firing_rule_lean_test.go` is repeated per node kind (confirmed present in
  `nodes/Time`, `nodes/TimeStart`, `nodes/pulse`, `nodes/PulseLeft`, `nodes/PulseRight`,
  `nodes/selectleft`, `nodes/selectright`, `nodes/input`, `nodes/holdflip`, `nodes/pacer`,
  `nodes/TimeEnd`, each ~107-112 lines). This boilerplate-per-kind is deliberate — each
  kind's firing rule is tested in isolation per the testing-shape doctrine (one goroutine
  asserts what it itself decided).
- Do NOT report duplication between node kinds on structural grounds. A structural-similarity
  score measures normalized SHAPE — identifiers and literals are stripped, which is what lets
  such a tool find renamed copy-paste, and also what makes it unable to tell "the same fact
  twice" from "two different facts that look alike". `nodes/selectleft` vs
  `nodes/selectright` are mirror-image logic (opposite select direction) and score near
  perfect while being deliberately OPPOSITE in behavior. **A high structural score between
  two node kinds is not evidence of duplicate behavior.**

  Two further reasons not to act on it here. Each kind's `RegisterBuilder` block states its
  own ports and injected fields; that repetition is what turned a renamed field from a
  silently-nil bug into a compile error, and the only genuinely universal lines are
  `n.Fire = a.Fire()` and `n.SpeedCh = a.SpeedCh()` (counted across all 11 kinds) — hoisting
  two lines behind a helper would make every kind's interface require reading a second file.
  And a duplication score can only ever recommend consolidation, never separation, so applied
  as a verdict it is a centralization ratchet pulling directly against the blast-radius
  audit. Read shape as a question, never as a finding.

## 2. Buffer-column lockstep is the intrinsic cost of the agnostic buffer

`Buffer/layout.go` → `Buffer/buffer_layout_gen.go` (Go, generated) +
`tools/topology-vscode/src/schema/buffer-layout.ts` (TS, generated) must agree column-for-
column. This is not accidental duplication to fix — it is the mechanism by which the binary
content buffer stays agnostic between the Go producer and the TS consumer. It is fully
guarded:
- `tools/check-generated.sh` — generated files match their generator.
- `tools/check-buffer-layout-parity.sh` — Go/TS column layouts agree.
- `tools/check-no-dead-buffer-column.sh` — no column is defined but unused.

Do not propose collapsing this lockstep; propose only guard gaps if the parity guards
above are found not to cover a specific new column.

## 3. Gap-numbered wire values are deliberate, never renumbered

`nodes/Wiring/input_codec.go` documents the input-record kind bytes as gap-numbered:
`save=4`, `raw-input=10`, `edit-update=22`, with kind bytes 20 and 21 explicitly left as
gaps ("kind bytes (20, 21) are left as GAPS below, never renumbered"). Likewise
`Buffer/frame_tags.go` numbers its four `BufBlockTag*` constants 4-7
(`BufBlockTagView=4`, `BufBlockTagEdgeStream=5`, `BufBlockTagNodeStream=6`,
`BufBlockTagInteriorStream=7`). Live wire values are never renumbered once assigned, even
if this leaves gaps — renumbering would silently break any code holding an old build's
binary framing. Do not report "unused/gap enum values" as a defect here.

## 4. Fingerprint strings duplicated in Go and TS are deliberate and effective

Several `*Fingerprint` string constants (e.g. `InputLayoutFingerprint` in
`nodes/Wiring/input_codec.go`, mirrored in
`tools/topology-vscode/src/schema/input-layout-gen.ts`; similar fingerprints for the
buffer layout in `Buffer/layout.go` / `tools/gen-node-defs/buffer_layout.go` /
`tools/topology-vscode/src/schema/buffer-layout.ts`) are long literal strings (each several
hundred bytes, several KB total across all of them) that encode the full shape of a wire
protocol in one line specifically so any drift between Go and TS trips a string-equality
test immediately (see e.g. `nodes/Wiring/input_codec_test.go`) instead of drifting silently
column-by-column. This is a deliberate, working anti-drift mechanism, not accidental
duplication — do not propose replacing it with something "less duplicated" unless the
replacement preserves the single-string-equality-check property.

## 5. `Wiring.Registry` is init-only self-registration, not shared mutable state

`Wiring.Registry` is a package-level map. This looks like a shared mutable map but is not a
concurrency hazard: the only writer is `Wiring.RegisterBuilder`
(`nodes/Wiring/build_args.go`), called exactly once per node package from that package's own
`init()`, and it panics on a duplicate key. All writes happen before `main` runs and before
any goroutine starts; after that it is read-only. Do not report this as a `sync`-free
shared-state violation.

(Superseded detail, kept because older commits and docs still name it: this used to be
`wire.KindRegistry` written by `wire.Register`, holding `func() any` constructors that
returned EMPTY structs for the central reflection pipeline to fill in. Both are gone —
`nodes/wire/registry.go` is now only a retirement note. The init-only property is unchanged;
only the writer's name and what it stores changed.)

## 6. Package-level var maps in `nodes/Wiring` are read-only dispatch tables

`nodes/Wiring/gesture.go` (`rawInputHandlers`, `hitClassifiers`), `gesture_graph.go`
(`commitEdges`, `applyAction`), `distance_groups.go` (`distanceGroupOrder`,
`distanceGroups`), and `port_wiring.go` (`speedChanFieldNames`) all declare package-level
`var` maps/slices. Each is a composite literal fully initialized at declaration (function
dispatch tables, static ordering lists) and is never written to after declaration — grep
confirms no assignment to these identifiers outside their `var` statement. These are
read-only dispatch tables, not shared mutable state; do not flag them alongside genuine
`sync.Mutex`/`atomic.`-style hazards.

## 7. DeltaA/B/C is established project vocabulary, not r/theta/phi hiding drift

`AbcDragLabel.tsx` (`tools/topology-vscode/src/webview/three/AbcDragLabel.tsx`) and an
`abc-drag` trace kind (`tools/topology-vscode/src/schema/trace-kinds.ts`) are real,
in-use vocabulary. A/B/C are **index deltas** (abc-index × step-constant — see
`memory/feedback_abc_times_constant_not_rederive.md`); r/theta/phi are the continuous
polar coordinates the layout ultimately resolves to. These are two different concepts at
two different layers, not a naming inconsistency where one is quietly standing in for the
other. Do not report DeltaA/DeltaB/DeltaC naming as drift or as an attempt to obscure
r/theta/phi.

## What still counts as a NEW finding

Anything not covered above — e.g. a fingerprint that's gone stale against its guard, a
"dispatch table" that gained a post-init mutation, a clone-family node whose *behavior*
(not just structure) has silently diverged from its sibling, an unguarded new lockstep
cluster, or any of the categories in the audit skill briefs that isn't one of the seven
items above. Verify against current code before reporting — this file itself only holds
as long as the code underneath it does.
