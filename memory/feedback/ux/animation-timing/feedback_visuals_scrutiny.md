---
name: Visual fixes need to scale and stay flexible
description: Editor visual fixes should use general mechanisms that scale across future audits, not point patches tuned to one observation
type: feedback
---

Visual fixes in the visual editor (tools/topology-vscode/) should be
built on general mechanisms that scale and stay flexible to future
visual audits, not point patches tuned to the current observation.

**Why:** repeated smoothness-audit iterations on pulse animation
(task/fix-pulse-overlap, 2026-05-03) showed that point-tuning a
constant or layering one-off conditionals lost ground to each next
observation. A general structure (uniform speed model, queueing,
preserve-progress on geometry change, single source of truth for
position) held up across multiple observation rounds; tuned-constant
approaches did not.

**How to apply:** when a visual issue surfaces, prefer the general
mechanism (clear invariant, parameterized model, structural fix)
over the smallest patch that addresses the immediate symptom. Even
if the user asks for a quick tune, flag if the underlying model is
the real lever. Expect any visual fix to be re-evaluated against
later observations, so leave the design open to that.
