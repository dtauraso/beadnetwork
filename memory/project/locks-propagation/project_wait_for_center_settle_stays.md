---
name: wait-for-center-settle-stays
description: waitForCenterSettle's silent 200ms timeout is a COSTED, DECLINED removal, not open work. Do not re-surface it. Its real deletion rides with the standing redesign where an edge carries its own length on its own EDGE frame.
type: project
---

**Decision (2026-07-31): `waitForCenterSettle` STAYS AS IS.** It returns silently on a 200 ms timeout with no signal it never settled. That is a known defect and it was costed and declined. It is not open work.

**Why it cannot simply be deleted.** It could be removed by measuring every pair's centers ONCE up front and computing all target positions from that snapshot — the wait exists only because `ApplyDistanceGroupTarget` re-reads centers mid-loop, so a pair aims from a node an earlier pair just moved (in "time", node 4 is the target of 2→4 and the source of 4→7/4→6).

But that changes two user-visible behaviours to buy one silent-timeout signal:

1. Chained pairs in "time"/"select" stop compounding down the chain on a ▲/▼ click.
2. The distance panel's number would have to come from the length dispatch REQUESTED rather than one it measures back, because `emitViewFrame` reads live centers and would then run before the movers commit.

Not worth it.

**The real deletion rides with the standing redesign** — an edge carrying its own length on its own EDGE frame — which removes the ordered loop and the wait together, and answers the VIEW-frame constraint properly.

Note it was deliberately EXCLUDED from the 2026-07-28 missing-upper-bounds sweep: it is bounded by time, so it is not a missing-bound case. Its defect deserves its own decision, and this is that decision.
