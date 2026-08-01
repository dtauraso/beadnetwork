# Plan — two clocks per bead, three channel sets

## THIS PLAN IS THE AUTHORITY

Where this plan and the existing code disagree, **the code changes**. The plan does not
bend to fit what is already there.

This is written first because it has already failed once, four times over in the same
session, each in the same direction — the shape already in the code won and the plan's
distinctive property was quietly dropped while its vocabulary was kept:

- per-edge bead SIZING varied bead size around a rounded `quantIR` instead of fixing where
  `count` came from;
- a CONFLICT VETO froze the node rather than letting per-edge CRUD be independent, because
  "the edges must agree" is the shape a central solver wants;
- a SUMMATION of per-bead verdicts into one displacement — a solver wearing addition;
- and ACTORS THAT MIRROR the central computation instead of replacing it (commit
  `100955e0`): `chainBeads()` kept computing `offsetAt(i)` and HANDED it to the beads, with
  the old inline math retained as a per-bead fallback. A replacement has nothing to fall
  back to and does not take the replaced thing's output as an input.

Every one of those preserved central computation over per-entity ownership — the exact
inversion this project exists to challenge (MODEL.md; `memory/feedback_derive_model_from_visual_spec.md`).

## THE REPLACEMENT IS NOT DONE UNTIL THE OLD PATH IS DELETED

No behavioural test can tell a replacement from a mirror — they render identically. So
completion is defined by SOURCE, and these are hard requirements, not preferences:

1. `offsetAt` **does not exist** in `nodes/Wiring/chain_beads.go`, and no per-index offset
   function is passed into the bead layer. A bead derives its own position from the
   broadcast transform it receives; it is not seeded with a value the caller computed.
2. **No fallback branch.** There is no "use the actor's position if valid, else compute it
   here" — the `else` at `chain_beads.go:420` and everything like it is gone.
3. `chainBeads()` **no longer computes bead positions.** It either reads what the beads
   already are, or it ceases to exist.
4. A **source guard** asserts 1-3, wired into the check suite, so this cannot silently
   regress to a mirror. Green stop-checks and a passing test suite are satisfied by the
   WRONG implementation; only a source check distinguishes them.

If any of 1-4 cannot be met, STOP and report which and why. Do not ship a layer beside the
old code and describe the fallback as a safety property.


## The model

A bead is driven by **two clocks at once**, over **separate sets of channels**:

| clock | rate | channel set | writes |
|---|---|---|---|
| **system** | CPU speed, in-pass | geometry channels | the bead's POSITION |
| **human-speed** (`MsPerTick = 16`, 62.5 ticks/s) | one tick per 16 ms | animation channels | the bead's ANIMATION state (lit, carried value) |

They never conflict, and not because they take turns — **they touch disjoint state**.
Position has exactly one writer, reached only over the geometry channels. Animation state
has exactly one writer, reached only over the animation channels. No field has two writers,
so there is nothing to lock and nothing to order.

Ownership + message passing, the same shape the rest of the network uses (guard:
`tools/check-no-network-locks.sh`, empty allowlist — no mutexes, no atomics).

### Modes, and a third channel set that switches them

A bead **runs on the human clock until it gets a message.** That is its resting mode:
checking whether its next animation tick is due, and animating when it is.

The mode is **one flag plus two messages**:

- **`dragging`** -> set the flag: the bead is on machine time, servicing its geometry
  channel and updating position at CPU speed.
- **`done dragging`** -> clear the flag: back to human speed, blocked on the tick channel,
  animating only when there is something in flight.

The flag is **local state owned by the bead's own goroutine** — written only by that
goroutine when it receives one of the two messages, read only by it. Nothing else touches
it, so it needs no lock and no atomic (guard: `tools/check-no-network-locks.sh`). The bead
is only ever in one mode, so the two clocks are never both driving it — a second,
structural reason they cannot conflict, on top of their state being disjoint.

The bead's own goroutine sits in one `select` over its channel sets — mode, geometry, and
tick — **with no `default:` case**. That distinction is load-bearing:

- `select` WITH `default:` is non-blocking, so a loop around it SPINS and burns a core.
- `select` WITHOUT `default:` BLOCKS — the runtime parks the goroutine on the channels'
  wait queues, takes it off the run queue, and wakes it when a sender arrives. Zero cycles
  while waiting. It is not checking anything; it is descheduled.

So the human tick must ARRIVE AS A MESSAGE on a channel, never be polled for. A
"has 16 ms elapsed yet" comparison cannot block, which forces a `default:`, which forces
the spin. One clock goroutine is the only thing in the system that waits on wall time, and
it SENDS; every bead and every mover blocks on messages.

Three channel sets: **geometry**, **animation/tick**, **mode**.

### The lifecycle — this is how both clocks AND idle CPU are had at once

1. **Idle.** Nothing is being dragged and nothing is in flight. Every bead is blocked in
   its `select`. Zero CPU. No polling, because there is nothing to poll for.
2. **A drag begins.** A **`dragging`** message wakes them all — every affected bead sets
   its flag and is now on machine time.
3. **Machine time.** Woken beads resolve position at CPU speed, paced only by pointer
   events (60-120/s, bounded by the input device — no free-running loop to throttle).
4. **Settle.** A **`done dragging`** message clears the flag: back to human speed, blocked
   again on the tick channel, animating only when there is something in flight.

**This is what per-bead goroutines buy.** Data cannot be asleep and cannot wake. A bead
that is a goroutine can be parked at zero cost and burst to machine speed on demand; a bead
that is a slice entry is recomputed by whoever owns the slice, on whatever clock that owner
happens to be running. The measurable claims are therefore: **idle CPU at zero**, and
**drag latency at machine speed** — both testable, and neither achievable today.

### The wake comes from the endpoint nodes, and reaches every bead at once

**Each node on either end of an edge owns a wakeup channel to all of that edge's beads.**
The node is the source of the wake; the beads on its edges are the receivers. A bead can be
woken by EITHER endpoint, since either end moving is a reason for it to be on machine time.

Make the wake **simultaneous, not a loop of N sends**: a bead blocks on receive from the
wakeup channel, and the node **closes** it. Closing wakes every goroutine blocked on it at
once — one operation regardless of how many beads there are, which is what "they get the
signal at the same time" requires. A channel can only be closed once, so each drag allocates
a fresh wakeup channel (a drag epoch) and closes it to start.

This also removes the 2N-sends cost noted below: the wake is O(1) to issue. The runtime
still makes N goroutines runnable, but nothing walks a list sending to each in turn, and no
bead is woken measurably later than any other.

Once woken, a bead does exactly two things: **move** (position, on the geometry channel,
machine time) and **send its colour** (lit / carried value, on the animation channel, human
speed). Those are the two disjoint pieces of state from the table above — one per clock.

Both endpoints own a wakeup channel because either end moving is a reason for the bead to
be on machine time — but **only one node is ever dragged at a time** (the gesture FSM holds
a single `dragNode`), so a bead is woken by exactly one end per drag. Do not build
reference counting, idempotence handling, or "wait for both ends to finish" logic: the
input model cannot produce that case.

### The flag is set ONCE per drag, not per pointer event

`dragging` fires at drag START and `done dragging` at drag END. In between, the beads stay
woken and simply receive geometry messages as pointer events arrive (~60/s). Do NOT toggle
the flag per event.

The cost difference is real at scale, and it is channel operations rather than arithmetic
that dominate. Per pointer event the position work is ~N float ops — for N = 1000 that is
single-digit microseconds against a 16.67 ms frame budget, about 0.05%, invisible. But a
wake/settle broadcast is 2N channel sends: at N = 1000 and ~1 us per send that is ~2 ms,
roughly 12% of the frame budget, paid on EVERY event if the flag toggles per event. Set
once per drag, the same cost is paid twice for the whole gesture instead of 60 times a
second.

So the human-visible claim is: a drag reacts within a frame with ~99.9% of the budget
spare, and the machine-time burst is not perceptible at human speed.

## No sleeping — the human clock CHECKS until it is time

**No `time.Sleep`, no `time.After`, no `time.NewTicker`**, no blocking timer on either
clock path. The clock **checks whether it is time to run yet**; if not, it does not run.
The decision is a comparison, not a suspension.

Beyond style: a sleep hands pacing to the runtime's timer heap, so the goroutine no longer
owns when it acts. A check keeps that local. It is also what lets the two clocks share a
bead — a goroutine parked in a sleep cannot service its geometry channel, so a sleeping
human-clock path would stall system-clock position updates. Checking never blocks the other
clock.

Guard it: a `Sleep`/`After`/`Ticker` on either clock path is a hard failure.

## Why two clocks at all

`nodes/wire/clock.go:39` names the existing one outright: "MsPerTick is the scale of the
**human-speed clock**". It exists so a person can watch a bead cross a wire — a presentation
rate. Geometry must never run on it: one propagation hop per tick costs 16 ms, so even
linear propagation is visible (40 beads = 0.64 s, 1000 = 16 s). That is what made the
original bead-chain edge take ~1.5 s at N~40 and get reverted.

**INVARIANT: no position update may be gated on the human clock, and no animation step may
run on the system clock.**

## Position updates: body force, not propagation

**Settled: a dragged chain must NOT visibly lag.** That makes position a body-force problem
— the node's motion applies to every bead at once, as gravity acts on every atom of a planet
simultaneously, with no information travelling along the chain.

The geometry channels carry the node's transform/anchors to every bead **in one hop**. Each
bead computes its own position directly from what it receives (dependency depth 1). No
hop-by-hop relaxation, no settling.

- **No diffusion.** A bead must never set its position from the average of its neighbours'
  positions. Midpoint averaging without momentum is the heat equation — information spreads
  as sqrt(rounds), so crossing N beads costs N^2 rounds. That rule IS the original defect.
- **No propagation hop per message round-trip.** A hop per goroutine round-trip (~100 ns -
  1 us) is the human clock in different clothes; the cost moves from the tick rate to the
  scheduler. One broadcast hop per event, not N.
- **Neighbour messaging is not the enemy.** Local, bounded fan-out is fine; GLOBAL scope —
  one place knowing all nodes, or all-to-all — is the thing to avoid
  (`memory/project_wire_is_straight_line_not_chain.md`).

## Animation: unchanged, on the human clock

Bead traversal and lighting keep running as they do now, over the animation channels.
`DwellTicksPerBead` and uniform pulse speed are untouched
(`memory/feedback_uniform_pulse_speed.md`: pulse speed is uniform across all wires).

## What this must NOT become

- **Not a global solver.** The bead-cell sphere-intersection solver was built and deleted.
- **Not one goroutine per bead per propagation step.**
- **Not TS-side geometry** (guard: `tools/check-no-webview-state.sh`).
- **Not per-edge bead sizing, and no arc/curve.** Both built and rejected; uniform bead size
  and uniform `wire.BeadStepR` spacing stay.

## Open before building

1. **Do beads become goroutines?** Today a bead is not an entity — `chainBeads` returns
   position slices, and MODEL.md has `PacedWire` as a passive delay queue stepped by its
   source node's goroutine, explicitly not a goroutine itself. This plan makes every bead a
   goroutine with three channel sets, so MODEL.md needs rewriting. The wake/settle
   lifecycle above is what requires it — data cannot park or wake.

ANSWERED — what per-bead goroutines buy: idle CPU at zero (blocked, parked) plus machine-
speed resolution during a drag (woken by broadcast). See the lifecycle section.

ANSWERED — does the existing clock path sleep: YES, and it must be converted.
`RealClock.SleepCycle` (`nodes/wire/clock.go:159`) blocks on `time.After(tickPeriod)`; it
is part of the `Clock` interface (`clock.go:66`) and every mover calls it each cycle
(`nodes/Wiring/node_mover.go:1039`). A mover parked in `time.After` cannot service any
other channel, which is exactly the stall this design must not have. Converting it into a
tick SEND from the single clock goroutine — with movers blocking on receive — is step 1 of
the work, not a rule for new code.

## Steps (once 1 is answered)

1. Write the two-clock invariant into MODEL.md; add a guard that fails if a position update
   is reached from the human-clock path or an animation step from the system-clock path.
2. Give the bead its three channel sets, with the state each owns structurally distinct —
   not several channels writing one struct field by convention, but a split the compiler can
   see. The goroutine is one `select` over mode channels, geometry channels, and a due-check
   of the human tick; human-clock mode is the resting state.
3. Geometry channels: one broadcast hop per event carrying the node transform/anchors; each
   bead computes its own position from it. No neighbour reads for position.
4. Animation channels: unchanged behaviour, human clock.
5. Correct `memory/project_wire_is_straight_line_not_chain.md` — its measurement (~1.5 s at
   N~40) is right, its conclusion ("don't re-propose the chain model") is attached to the
   wrong cause. The causes were human-clock gating and the momentum-free midpoint rule.

## Tests

- **Clock separation**: a position update completes with the human clock STOPPED; an
  animation step does not advance position.
- **Mode switching**: a bead rests in human-clock mode; a mode message moves it to
  system-clock mode and another returns it. Never in both modes. A switch loses neither a
  pending animation tick nor a pending geometry update.
- **No human-clock coupling in geometry**: resolving N beads' positions takes one pass for
  N = 40 and N = 1000 — fail if it needs more than one human tick.
- **One hop, not N**: a node move reaches every bead in a single broadcast; hop count does
  not grow with N. This is the test that would have caught the original defect.
- **No diffusion**: a bead's position must not be a function of its neighbours' positions.
- **Disjoint writers**: no field is written from both channel sets.
- **No sleeping**: source guard against `Sleep`/`After`/`Ticker` outside the single clock
  goroutine, plus behaviourally — a bead waiting for its next tick still services its
  geometry channel immediately.
- **No spinning**: no `select` on a bead or mover hot loop carries a `default:` case. Assert
  by source guard — this is the difference between parked at zero CPU and burning a core,
  and it is invisible in any behavioural test.
- **Idle costs nothing**: with no drag and nothing in flight, beads are blocked and consume
  no CPU. Assert that no bead goroutine is runnable in that state.
- **Wake and settle**: a `dragging` message sets every affected bead's flag; `done dragging`
  clears it. Assert both transitions, and that a bead woken mid-animation loses neither its
  pending tick nor its pending geometry update.
- **`done dragging` is not optional**: a bead that never receives it must not be left on
  machine time burning CPU. Assert that a drag ending by any path — including one abandoned
  without a clean end event — settles every bead it woke.
- **The flag is set once per drag**: assert exactly one `dragging` and one `done dragging`
  per gesture, regardless of how many pointer events it contains. A per-event toggle passes
  every behavioural test while paying 2N channel sends 60 times a second.
- **The wake is one operation, not N**: assert the node issues a single close rather than
  iterating its beads. A send-loop is behaviourally identical and scales with N.
- **Either endpoint can wake**: dragging the SOURCE node wakes the edge's beads, and so does
  dragging the TARGET node. One test each — not a both-at-once test, which the single-drag
  gesture model cannot produce.
- **Frame budget**: a pointer event's position work for N = 1000 completes well inside one
  16.67 ms frame. Measure it rather than assume it.
- Existing chain-bead and bead-CRUD tests keep passing, unweakened.

## Verify

`bash scripts/stop-checks.sh` — read its **stdout**, it always exits 0. Clean == empty
stdout. Never run the sim in the foreground.
