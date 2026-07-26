---
name: audit-all
description: Run all three repo audits (blast-radius, priors-fit, grep-load) in parallel as read-only Explore subagents, then consolidate into one deduped, ranked findings table separating real-and-actionable from deliberate-and-structural.
---

Run all three repo audits **in parallel** and present ONE consolidated table. Read-only —
do NOT edit any files. The three audits overlap; the value of running them together is the
deduped cross-audit synthesis, not three separate reports.

Spawn **three `Explore` subagents concurrently** (one message, three tool calls), each with
one audit brief below. Scope every subagent out of `node_modules,out,.git,handoff-archive,memory`
and tell each to grep-first and return a concise categorized findings list (file:line,
High/Med/Low rank, no fixes). Full briefs live in the sibling skills — keep each subagent's
task faithful to them:

1. **blast-radius** (`.claude/skills/audit-blast-radius/SKILL.md`) — structural cost: shared
   mutable state (`sync.Mutex`/`sync.RWMutex`/`atomic.`/post-init package vars/shared maps),
   wide coupling (high-fan-in packages, files importing many packages, god-modules >400 lines),
   and change-path centrality (lockstep schema/codec/layout edit clusters).
2. **priors-fit** (`.claude/skills/audit-priors-fit/SKILL.md`) — cognitive cost: non-idiomatic
   constructs (magic column indices, gap-numbered enums, off/len encodings, fingerprint-parity
   dances), concept-encoding mismatch (one concept spread across many files that must agree),
   and naming/vocabulary drift (grep-verified, not prose-trusted). Have this one read CLAUDE.md
   + MODEL.md first for the intended model.
3. **grep-load** (`.claude/skills/audit-grep-load/SKILL.md`) — verification cost (run this
   subagent on model: haiku): string/key duplication across Go↔JSON↔TS boundaries, doc claims
   that could drift, runtime-only validation a parser could catch earlier, and files claiming
   to be generated but hand-edited.

When all three return, present **one** table, deduped across audits (the same finding often
surfaces in two audits — merge it), grouped and ranked. Then split the findings into:

- **Real & actionable** — a genuine defect or un-guarded hazard. For each, name the concrete
  fix and whether a `check-*.sh` guard can lock it (fix-order is code-first: enforce in a
  guard before adding prose).
- **Deliberate & structural** — the intrinsic cost of the architecture (e.g. the agnostic
  buffer's schema lockstep), not a defect. Record as intentional, do not "fix."

Verify before recommending: an audit surfaces *candidates*. Confirm each against current code
(grep / read the site) before calling it real — several will not survive contact with the
code. Do NOT auto-fix; the user picks what to act on.
