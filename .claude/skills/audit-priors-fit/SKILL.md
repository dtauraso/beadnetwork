---
name: audit-priors-fit
description: Run the read-only priors-fit code-smell audit (non-idiomatic constructs, concept-encoding mismatch, vocabulary drift) as an Explore subagent and return a ranked findings table.
---

Launch a **read-only** priors-fit code-smell audit of this repo (Go network + TS webview)
via the `Explore` subagent. Do NOT edit any files.

Goal: find "priors-fit" problems — idiosyncratic constructs where a concept is encoded in a
form an LLM can't predict cheaply, causing generate-then-back-out (thrash). This feeds the
token-reduction program (target ~5× less AI usage per feature).

Give the subagent this task (read CLAUDE.md + MODEL.md first for the intended model, then
grep-first; scope out `node_modules,out,.git,handoff-archive,memory`):

1. **Non-idiomatic / surprising constructs** — hand-rolled stdlib things, unusual
   concurrency, bespoke binary encode/decode, magic numeric column indices, gap-numbered
   enums/kind-bytes, off/len encodings, fingerprint-parity dances. file:line + why it thrashes.
2. **Concept-encoding mismatch** — a domain concept spread across many files that must agree
   (parity guards, generated-from-source pairs, dual Go+TS schema) so the "one obvious form"
   is actually several coupled forms. List the clusters.
3. **Naming / vocabulary drift** — dead vocabulary, misleading comments, stale doc claims
   about code. **Verify each flagged item actually exists right now with grep** and report
   file:line (line pointers in tracker prose drift — trust grep, not the prose).

Rank each High/Med/Low thrash cost. Concrete paths. No fixes — findings only, as raw
material for a categorized table.

Note: `tools/check-comment-vocab.sh` already guards retired comment vocabulary
deterministically — this audit finds NEW drift to add to that guard's token list.

When it returns, present the findings as a table by category and rank.
