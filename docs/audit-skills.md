# The three audits

Three read-only skills survey this repo for the cost of changing it. They serve one program:
**less AI usage per feature** — shrink blast radius, fit the model's priors, then lock each
win with a guard so it cannot regress.

None of them edits anything. Each returns candidates; a candidate is not a finding until
someone confirms it against current code. Several never survive that contact.

`audit-all` runs the first two in parallel and consolidates.

## The split: three costs, not three severities

Each audit measures a different way the codebase can be expensive, and the three do not
substitute for one another.

| Audit | The cost it measures | What that cost looks like |
|---|---|---|
| **blast-radius** | *structural* — how much of the system you must load to change one part safely | shared mutable state, god-modules, lockstep edit clusters |
| **priors-fit** | *cognitive* — how often a model generates the predictable form, then backs it out | non-idiomatic constructs, one concept spread across coupled forms, dead vocabulary |
| **grep-load** | *verification* — how often you must grep to answer "is this claim still true?" | duplicated strings, doc claims about code, runtime-only checks, fake generated files |

A file can be cheap on one axis and ruinous on another. A tiny, idiomatic file that four
other files must be edited in lockstep with is fine by priors-fit and terrible by
blast-radius.

## blast-radius — structural cost

Three categories:

1. **Shared mutable state** in Go — `sync.Mutex`, `sync.RWMutex`, `atomic.`, package-level
   `var` written after init, maps written by several goroutines. The doctrine is *ownership
   replaces locking*, so each hit is a candidate violation, reported with what it guards and
   how many call sites touch it.
2. **Wide coupling** — packages with high fan-in, files importing many packages, single
   files over 400 lines doing several jobs.
3. **Change-path centrality** — the files CLAUDE.md's primitive-landing rule forces you to
   edit together (schema / codec / layout).

## priors-fit — cognitive cost

Reads CLAUDE.md and MODEL.md first, for the intended model to measure against. Three
categories:

1. **Non-idiomatic constructs** — hand-rolled stdlib, bespoke binary encode/decode, magic
   numeric column indices, gap-numbered kind bytes, off/len encodings, fingerprint-parity
   dances. Reported with *why it thrashes*, not just that it is unusual.
2. **Concept-encoding mismatch** — one domain concept spread across files that must agree,
   so the "one obvious form" is really several coupled forms.
3. **Vocabulary drift** — dead names, misleading comments, stale doc claims. Every item must
   be grep-confirmed to exist right now: line pointers in prose drift, so this category
   trusts grep and never the prose it is auditing.

`tools/docs/check-comment-vocab.sh` already enforces retired vocabulary deterministically.
This audit's job is to find NEW drift worth adding to that guard's token list — which is the
shape the whole program wants: the audit finds it once, the guard holds it forever.

## grep-load — verification cost

On demand only, and the reason is measured rather than assumed: a completed run cost **85
tool calls and ~98k tokens to return one net-new finding**, because its categories now
overlap the `check-*-parity.sh` guard suite and baseline sections 2–4. Run it when someone
asks, or after a change that plausibly opened a new duplication surface a parity guard does
not yet cover — a new Go↔TS boundary, a new wire-protocol codec.

Four categories: string/key duplication across boundaries; documentation claims about code
that could drift; runtime-only validation a parser could have caught earlier; and files
carrying a "GENERATED — DO NOT EDIT" header that nothing actually regenerates.

Findings are ranked by *leverage* — how many drift hazards one fix eliminates. Removing a
duplicated string that any change might need to update ranks high; fixing a lone instance
with no wider pattern ranks last, though still worth doing.

## What happens to a finding

Consolidated results split in two, and the split is the point:

- **Real and actionable** — a defect or an unguarded hazard. Name the concrete fix and
  whether a `check-*.sh` guard can lock it. Fix-order is **code first**: enforce it in a
  guard before writing prose about it.
- **Deliberate and structural** — the intrinsic cost of this architecture, not a defect. Say
  so, and leave it alone.

Every run starts from scratch. There is no list of settled findings to read first, so the
same deliberate-and-structural facts surface again each time and have to be judged again
each time. That is the accepted cost of not keeping a record that goes stale faster than the
code it describes.

Nothing is auto-fixed. The user picks.
