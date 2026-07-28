---
branch: task/runtime-invariant-assertions
---

# Plan: the first runtime assertion

Companion to `runtime-invariant-assertions.md` (that file states the *why* and the scope
rule). This file is the concrete sequence. Prerequisite met:
`task/synctest-deterministic-tests` landed on main (Go 1.25, clock tests exact), so the
ordering note in the parent doc is satisfied and assertions can be verified under a
deterministic scheduler rather than a hopeful one.

## House idiom — follow it, do not invent a helper

Three assertion sites already exist and agree on a shape:

```go
panic("allocateWires: two edges target " + destKey +
      " — validateNoFanIn should have rejected this fan-in at parse")
```

- `nodes/Wiring/build.go:178` — fan-in reached wire allocation
- `nodes/Wiring/input_codec.go:114` — fingerprint missing a required token
- `nodes/wire/registry.go:16` — duplicate kind registration

The shape is: **site name, the invariant violated, and what upstream should have caught
it.** No `assert()` helper exists and none should be added — a helper would obscure the
message, which is the part that makes a panic cheap to act on. `build.go:178` is already
a runtime invariant assertion in this exact spirit, so this branch extends an existing
practice rather than introducing one.

## What the first assertion actually is

The parent doc's candidate was "an emit lands on its owner's stream." Checking the code
first changed the shape of that, and the difference matters:

**Not assertable as written.** `nodeMover.streamOut` (`nodes/Wiring/node_mover.go:232`)
is a per-mover field — each mover structurally holds only its own fd, so "wrote someone
else's stream" cannot be expressed by writing through `m`. And the stricter reading —
"only this mover's goroutine calls `writeStreamFrame`" — needs a goroutine identity Go
does not cheaply expose. Asserting either would be theatre.

**Assertable, and the real content of the invariant.** Two facts line up:

- `nodeMover.nodeRow` (`node_mover.go:236`) is this node's stable buffer NODE-ROW index.
- `wire.RowEvent.NodeRow` (`nodes/wire/owner_events.go:20`) names the row an event is
  *about*.

So at `nodeMover.writeStreamFrame` (`node_mover.go:733`) the check is: **every event this
mover appends to its own stream must name its own `nodeRow`.** A mover carrying another
node's event out on its own dedicated stream is precisely the cross-owner write
`memory/feedback_no_single_writer_bridge.md` and `memory/feedback_per_goroutine_bridge.md`
forbid, and it is a single int32 comparison at one boundary — cheap enough to leave on.

### Confirm before asserting

`RowEvent` also carries `TargetRow`, `TargetPortRow`, and `EdgeRow`. Those are almost
certainly *references* to other rows (legitimate — an edge event names both ends), not
ownership claims. Only `NodeRow` should be constrained. **Verify that reading against the
real event producers before writing the assertion** — if some Kind legitimately emits a
foreign `NodeRow`, the assertion is wrong as stated and needs narrowing to the Kinds where
ownership genuinely holds. Do not assert first and narrow after: a panic that fires on
correct behavior trains the reflex to disable it.

## Steps

1. Enumerate every producer of the `events` slice reaching `writeStreamFrame` and record,
   per `Kind`, whether `NodeRow` is an ownership claim or a reference.
2. Add the assertion at the top of `writeStreamFrame`, in the house idiom, covering the
   Kinds step 1 confirmed. Name the offending row and the owner in the message.
3. Make it fail once, deliberately — emit a foreign `NodeRow` and confirm the panic fires
   and reads clearly (`memory/feedback_check_the_signal_the_check_emits.md`).
4. `bash scripts/verify.sh` clean.
5. Commit alone, mapping the assertion to the doctrine it encodes in the message.

## After the first one

Re-evaluate rather than batch. The parent doc's remaining candidates each need the same
"check the code before asserting" pass the first one just got:

- **simtime monotonic per node** — likely already structural: `RealClock.Tick()` is a pure
  function of `time.Now()` and speed history, so it cannot go backward by construction.
  Confirm before spending an assertion on it.
- **edge has exactly two endpoints matching its channel name** — the channel-naming
  convention is called load-bearing in CLAUDE.md; check whether anything reads the name
  back at runtime, since an unparsed name cannot be asserted against.
- **buffer layout parity at the point of write** — `check-buffer-layout-parity.sh` covers
  the static half; the runtime half is a real gap.

The edge equivalent (`edgeMover.writeStreamFrame`, `nodes/Wiring/edge_mover.go:235`) is
the obvious second site once the node one is proven.

## Done when

- Every added assertion traces to an existing guard, `memory/` entry, MODEL.md, or
  CLAUDE.md — the mapping stated in the commit message.
- Each has been made to fail once deliberately.
- `bash scripts/verify.sh` is clean.
