---
name: project-v0-cost-calibration
description: Post-Phase-5 cost ratios size the generation step only — post-v0 rearchitecting cost ~$1.8k in API usage, far more than any phase estimate
metadata:
  type: project
---

# v0 Cost Calibration

Phase 5 came in at ~$5.65 against a $110 estimate (~5% of midpoint, ~18× overestimate).
Cap-hit estimation retired post-Phase 5; the unit was calibrated against an older model
and a less mature codebase.

## Risk band ratios (post-Phase-5 recalibration)

| Phase type              | Ratio of original estimate |
|-------------------------|---------------------------|
| Mechanical              | ~10%                      |
| Hardening               | ~12%                      |
| Refactor / exploratory  | ~15–20%                   |

Mechanical phases run roughly an order of magnitude under original budget with
Opus 4.7 + existing harness/adapter/save infrastructure.

Hardening phases have more codebase exploration than pure-function authoring,
landing slightly above mechanical.

Refactor and exploratory phases carry wider risk bands: the Phase-5 efficiency
factor may not generalize fully to less-scoped work.

## What these ratios do NOT cover

**Rearchitecting the system several times after v0 was done cost ~$1.8k in API usage.** That is
two orders of magnitude above any single phase figure above, and none of the ratios
predict it.

The phase numbers measure the **generation step** — writing the code for a scoped,
already-specified phase. They do not measure the work that follows: rearchitecting,
and rearchitecting again. Reading ~$5.65 as "what this cost" is reading the cheapest
step and calling it the total.

## Usage

Apply these ratios when sizing the generation step of a scoped phase. Cap-hit column
is no longer load-bearing.

Do NOT use them to size a project, a redesign, or anything unscoped — the ~$1.8k
figure is the relevant anchor there, not the per-phase ratios. When quoting a cost,
say which of the two it is.
