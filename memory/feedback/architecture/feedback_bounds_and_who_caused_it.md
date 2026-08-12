---
name: bounds-and-who-caused-it
description: Whether a bound panics depends on WHO caused reaching it — a code bug panics, designed backpressure is unasserted, legitimate large input reports loudly without crashing. Also: a diagnostic naming the wrong cause is worse than a vague one, and "derive it from topology" is right for counts and wrong for depths.
type: feedback
---

Three rules from the 2026-07-28 explicit-upper-bounds sweep, each with its at-the-bound decision written down rather than applied as a blanket policy.

**Panic is not the answer everywhere; the dividing line is who caused it.**

- `maxPendingEvents`, `maxInflightBeads`, `maxPendingSends` PANIC — reaching them means a code bug.
- `moverInboxDepth` is named but deliberately NOT asserted — filling a mover inbox is the DESIGNED backpressure.
- `MAX_EDGE_STREAMS` overflow reports loudly WITHOUT crashing — a large topology is legitimate input.

**A diagnostic that names the wrong cause is worse than a vague one.** The first `maxInflightBeads` message blamed a destination that stopped draining `outCh`. True, but a SOURCE outpacing the wire reaches the same bound — that message would send a reader hunting a consumer that is working fine. Every such message must name every cause that applies.

**"Derive it from topology" is right for counts and wrong for depths.** The mover inboxes' COUNT is structural (one triple per edge). Their DEPTH is a gesture-rate queue and is not in any saved file. A plan promised a derivation it could not deliver; the constant now says outright that 8 is chosen.

**A bound on one side of a two-sided protocol is a silent failure.** `splitFrames` read `frameLen` off the wire with no maximum while Go enforced one on the same protocol. Now guarded against re-diverging.

See [[no-atomics-are-defects]], [[go-vs-coordinator-bias]].
