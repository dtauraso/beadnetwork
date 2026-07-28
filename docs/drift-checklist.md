# Drift checklist

Run this before declaring agent/model behavior healthy, or when "the AI is getting worse."
This is a **periodic audit**, not per-session context — it lives here rather than in
CLAUDE.md so it costs nothing on turns that aren't running it.

(Borrowed, ECC agent-architecture-audit — keep the questions, not the rest.)

Each question is tagged with how it's checked: **[guarded]** = a deterministic
`tools/check-*.sh` enforces it (run automatically by stop-checks); **[partial]** =
only slices are mechanizable; **[manual]** = inherently behavioral, no repo proxy exists.

1. Can the model skip a required step/tool and still answer? → not code-gated. **[partial —
   folds into #6: a required step with no code gate]**
2. Does old conversation content appear in new turns? → context contamination. **[guarded:
   `check-no-state-cache.sh` forbids the handoff/continuation snapshot files that
   reintroduce stale state — the mechanism, since the turns themselves aren't visible]**
3. Is the same info in CLAUDE.md AND `memory/` AND history? → context duplication.
   **[partial: `check-doc-drift`/`check-dead-doc-tokens`/`check-doc-citations`/
   `check-doc-symbols` cover slices; substantive-dup detection is false-positive-prone
   (intentional index/pointers), so stays manual]**
4. Does a second pass silently rewrite the answer before delivery? → hidden repair loop.
   **[guarded: `check-hooks-allowlist.sh` pins every `settings.json` hook to a known
   check/reminder script — a new hook that could rewrite output fails until reviewed]**
5. Does output differ between what was generated and what's delivered? → rendering
   corruption. **[manual — purely the harness delivery pipeline; no repo artifact evidences
   it]**
6. Are "must do X" rules only in prose, never enforced in code? → discipline failure.
   **[partial — the prose→guard mapping isn't mechanical; inventory the must/never lines by
   hand and confirm each has a guard/hook or is deliberately prose-only]**
7. Can the agent's own monologue become persistent `memory/`? → memory poisoning.
   **[guarded: `check-memory-hygiene.sh` — every `memory/*.md` must have valid frontmatter
   (name/description/type) and a `MEMORY.md` index entry, so malformed monologue can't
   silently persist]**

Fix order is code-first: enforce in code (a guard/hook) before adding more prose. Of the 7,
**3 are guarded** (#2/#4/#7), **3 are partial/manual** (#1/#3/#6), and **#5 is
inherently behavioral** — run those four by hand.

## Context cost

Question #3 has a cost dimension as well as a correctness one: duplicated doctrine is paid
for on every turn it loads. Region-specific detail belongs in `.claude/rules/` with `paths:`
frontmatter (loads on demand); only cross-cutting invariants where drift is expensive belong
in root CLAUDE.md, which loads in full every session and is re-injected after `/compact`.

Nested rules and subdirectory CLAUDE.md files are **not** re-injected after compaction —
they reload the next time a matching file is read. That is why load-bearing invariants stay
in the root file.
