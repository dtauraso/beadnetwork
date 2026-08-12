---
name: measure-bodies-not-signatures
description: A decomposition decline must name a mechanism the compiler or a rule enforces. Seven declines were overturned on task/god-objects for describing code instead of pinning it — signature-mentions-a-hub, impure-top-level, mutually-exclusive-branches, no-seam. Also ask "what are the phases?", not only "what is pure?", and measure code lines brace-to-brace.
type: feedback
---

**Rule:** a decline must name a mechanism the compiler or a stated rule enforces. Seven declines were overturned on `task/god-objects` because each DESCRIBED the code instead of PINNING it.

**Failure shapes, all seen:**

- *"the signature takes `*moverRegistry`"* — four functions declined this way used it for a single `mr.centerOfNode` call, and one never referenced it at all. All four moved by passing `centerOf func(string)(wire.Vec3,bool)` as a bound func value, the pattern `md.mr.enqueueFor(ng)` already used.
- *"impure — its top level sends on channels"* — the body was 265 lines, ~190 of them arithmetic on locals. Split into three pure phases.
- *"mutually-exclusive `if...return` branches"* — that is why ten branches were trivial to extract, not a reason to keep them in one function.
- *"no seam"* about a loop with five impure operations — measured, 24–55 extractable lines sat between each pair.
- *"reads `mr.nodeGeoms`"* — true of 1 statement out of 13.
- Correct analysis attached to the wrong goal: `buildDeps` embedding `*moverRegistry` genuinely blocks decoupling node kinds, and says nothing about whether files can move.

**Not a decline:** cohesive, one concern, no seam, hot path, entry point, impure, already small, it's JSX.

**A decline:** an import cycle `go build`/`tsc` reports; a write to another package's unexported field; an ordering invariant a split would reorder; a guard forbidding the placement; a hardcoded path a parser depends on; shared live state between the parts (`pole-singularity.html`'s two tabs share a slider value).

**Ask two questions, not one.** "Is there pure computation to lift?" does not reach a 190-line method. "What are the sequential phases?" needs no purity — `buildFromSpec` is a short orchestrator over three impure phases, and that shape resolved `run()` 190→18, `NewMoveDispatch` 184→26, `main()` 164→27, `LoadTree` 149→47.

**Measure CODE lines, brace-to-brace.** This repo ran ~43% comments before the strip; total-line counts overstate function size badly, and awk scans that count to the next `func` report file headers as functions. Six measurements were wrong that way in one session.

See [[guards-satisfied-by-text]].
