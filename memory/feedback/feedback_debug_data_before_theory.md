---
name: debug-data-before-theory
description: On a live editor/geometry bug, get runtime data (bisect a known-good build, instrument, repro) BEFORE hypothesizing a mechanism. Theorizing first burned many turns and produced two confidently-wrong fixes.
type: feedback
---

**Rule:** For an interactive/geometry/camera bug reported live, the FIRST move is data, not a theory: (1) check the `.probe` error logs; (2) if it might be a regression, `git checkout` a known-good pre-change build and A/B it (bisect the actual commit); (3) instrument the exact value in question with a breadcrumb and have David repro. Only propose a mechanism AFTER the data points at one. Do not build a fix on an unconfirmed mechanism.

**Why:** Reasoning-first on runtime behavior is high-confidence and usually wrong, and it compounds — a wrong theory leads to a wrong fix that can make things worse. Concrete failure (2026-07, camera regression): I pattern-matched "eventually-consistent mirror → lag," then "raw center vs snapped center," and shipped a fix on the second guess that made rotation WORSE — because the old atomic actually stored the RAW center too, so the value never changed. The real signal came only from a breadcrumb (mirror centers were perfectly stable, exonerating the theory) and a commit bisect (zoom = pre-existing at session start; rotation "worse" = accumulated running-process state across many reloads, not code). Both correct conclusions came from data; every intermediate theory was wrong.

**How to apply:** When tempted to explain a live bug, stop and ask "what number would confirm this, and can I log it or bisect for it?" A 30-second checkout of a pre-session build settles regression-vs-pre-existing better than any amount of static analysis. This is the repo's own guidance ([[feedback_runtime_breadcrumbs_beat_static_analysis]], [[feedback_runner_errors_probe_first]], [[feedback_cost_overruns]]) — I violated it and it cost real cycles.
