---
name: audit-all
description: Run the repo audits (blast-radius, priors-fit; grep-load on demand) in parallel as read-only Explore subagents, then consolidate into one deduped, ranked findings table separating real-and-actionable from deliberate-and-structural.
---

**First, read `docs/audit-baseline.md`** — it is the permanent record of findings already
judged deliberate-and-structural (node-kind clone families, buffer-column lockstep,
gap-numbered wire values, fingerprint-string duplication, `wire.KindRegistry`, read-only
dispatch-table var maps, DeltaA/B/C vocabulary). Do not have any subagent re-report these;
each brief below tells its subagent to read it too. This file exists precisely so audits
don't re-derive the same permanent architectural facts from scratch every run.

Run the audits **in parallel** and present ONE consolidated table. Read-only — do NOT edit
any files. The audits overlap; the value of running them together is the deduped
cross-audit synthesis, not separate reports.

By default, spawn **two `Explore` subagents concurrently** (one message, two tool calls),
each with one audit brief below. Scope every subagent out of
`node_modules,out,.git,handoff-archive,memory` and tell each to grep-first and return a
concise categorized findings list (file:line, High/Med/Low rank, no fixes). Full briefs
live in the sibling skills — keep each subagent's task faithful to them:

1. **blast-radius** (`.claude/skills/audit-blast-radius/SKILL.md`) — structural cost: shared
   mutable state (`sync.Mutex`/`sync.RWMutex`/`atomic.`/post-init package vars/shared maps),
   wide coupling (high-fan-in packages, files importing many packages, god-modules >400 lines),
   and change-path centrality (lockstep schema/codec/layout edit clusters).
2. **priors-fit** (`.claude/skills/audit-priors-fit/SKILL.md`) — cognitive cost: non-idiomatic
   constructs (magic column indices, gap-numbered enums, off/len encodings, fingerprint-parity
   dances), concept-encoding mismatch (one concept spread across many files that must agree),
   and naming/vocabulary drift (grep-verified, not prose-trusted). Have this one read CLAUDE.md
   + MODEL.md first for the intended model.

**grep-load is on-demand, not run by default.** Its four categories (string/key duplication,
doc-claim drift, runtime-only validation, generated-file drift) now largely overlap with the
`tools/check-*-parity.sh` guard suite and with `docs/audit-baseline.md` sections 2-4, so a
full run mostly re-confirms what's already guarded. Measured cost on a completed run: 85
tool calls and ~98k tokens to return one net-new finding. Only spawn it — via
`.claude/skills/audit-grep-load/SKILL.md` — when the user explicitly asks for a grep-load
pass, or after a change that plausibly opened a new duplication surface (e.g. a new
Go<->TS boundary, a new wire-protocol codec) not yet covered by a parity guard.

When the subagents return, present **one** table, deduped across audits (the same finding
often surfaces in two audits — merge it), grouped and ranked. Then split the findings into:

- **Real & actionable** — a genuine defect or un-guarded hazard. For each, name the concrete
  fix and whether a `check-*.sh` guard can lock it (fix-order is code-first: enforce in a
  guard before adding prose).
- **Deliberate & structural** — the intrinsic cost of the architecture (e.g. the agnostic
  buffer's schema lockstep), not a defect. Record as intentional, do not "fix." If it is a
  genuinely new deliberate/structural finding not already in `docs/audit-baseline.md`,
  add it there in the same pass so future audits don't re-derive it.

Verify before recommending: an audit surfaces *candidates*. Confirm each against current code
(grep / read the site) before calling it real — several will not survive contact with the
code. Do NOT auto-fix; the user picks what to act on.
