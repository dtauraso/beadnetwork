---
name: no-atomics-are-defects
description: In the node network an atomic (like a mutex) is a shared-state DEFECT, not a tool. Correct = single-goroutine ownership + message-passing, zero shared memory. Enforced by check-no-network-locks.sh (empty allowlist).
type: feedback
---

**Rule:** In the concurrent node network (`nodes/`, `Buffer/`, `Trace/`) there is ZERO shared mutable state. Each piece of state is owned by exactly one goroutine; anything another goroutine needs is SENT to it as a value, not read from a shared location. Under this model a synchronization primitive is a **defect marker**: if code reaches for `sync.Mutex`/`RWMutex` OR `sync/atomic`, it shared something it should have owned. Do not add either "for safety" or "for a fast lock-free read" — restructure so there is no cross-goroutine read at all.

**Why:** "Correct sharing via an atomic" is correct only at the Go-memory-model level; at this system's substance level it is still *sharing*, and the whole model is ownership-replaces-locking (MODEL.md). An atomic.Pointer publish-snapshot is a HALF-measure — it removes the lock but keeps a shared cell read by many. The strictly stronger position is: own the state, push updates to owned copies. The read-back trap that makes people reach for atomics (a consumer needing many nodes' state on demand) is a consequence of the *arrangement*, not a law of the domain — change the arrangement.

**How to apply:** `check-no-network-locks.sh` enforces this with an EMPTY `ALLOWED_ATOMIC` allowlist (mutexes forbidden outright; a new atomic fails the build). If you think you need one, that's the signal the ownership model broke — find the missing push. Worked example (2026-07): the three network atomics (msgTap, wire-segment snap, centerSnap) were all removed — msgTap→per-mover-owned tap, snap→single-threaded test, centerSnap→per-node channels feeding an owned camera mirror + neighbor push to owned `partnerCenters`. All three were `-race` clean with no behavior change. Related: [[feedback_no_single_writer_bridge]], [[feedback_per_goroutine_bridge]], [[feedback_node_model_not_networking_handshake]].
