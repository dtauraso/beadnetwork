---
name: guards-satisfied-by-text
description: Five guards were found checking something weaker than they claimed — a grep for a NAME that a call site satisfied, a count of FILES while claiming call SITES, a scan reading test fixtures as the decoder, a comment counted as a production consumer, a hardcoded single file. All looked green. Verify a guard fails, don't trust that it passes.
type: feedback
---

**Rule:** a guard that text-matches is satisfied by text, not by the property it names. Five were found this way on `task/god-objects`, every one green beforehand:

- `check-stream-kind-ts-parity.sh` — grepped for `handle<Kind>Fd(` anywhere under the ext-host tree. Renaming only the DEFINITION left it green, because a call site still contained the literal. Fixed by anchoring to a line-start definition.
- `check-no-dead-buffer-column.sh` — its consumer-scan counted a COMMENT naming a reader as a production consumer. `readEdgeSelected` had been "consumed" for months by the sentence documenting its own exemption — and that sentence claimed an `ALLOWED_DEAD` entry that was never made (the array was empty). Fixed by stripping comments before building the corpus; the column was then genuinely dead and deleted.
- `check-input-attr-dispatched.sh` — scoped its decoder search to the dispatch table's own package, which stopped containing the decoder when a refactor split them. It had been reading TEST FIXTURES as if they were the decoder. Fixed by searching repo-wide.
- `check-breadcrumb-label-registered.sh` — hardcoded `Trace/Trace.go`; splitting that file would have blinded it. Fixed to scan `Trace/*.go`. See [[guards-hardcoding-single-file-break-on-split]].
- `check-no-node-node-polar.sh` — uses `grep -rl`, so it counts FILES while its message claims "exactly one call SITE". A second summation in the same file passes. Left as-is: the drift it exists to stop is the summation escaping to another module, which it does catch — but its wording overstates it.

**Practice:** before trusting a guard, break the thing it protects and confirm it fails by name. A guard that cannot fail reads exactly like one that passes. `check-message-kind-parity.sh` is the counter-example worth copying — it refuses a vacuous pass on an empty extracted set and said so loudly when a comment block it parsed was removed.

**Comments can be configuration.** The comment strip broke three guards whose INPUT was a comment: a `MSG_TYPES` list parsed as data, `# PLACEMENT:` lines (63, read by `check-placement-declared.sh`, which fails if a guard declares none), and the dead-column consumer scan. Distinguish commentary from configuration before removing either.
