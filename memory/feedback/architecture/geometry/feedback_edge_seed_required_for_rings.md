---
name: feedback-edge-seed-required-for-rings
description: Ring startup deadlock is broken by a dedicated Input bootstrap node wired via a real edge into the feedback port; the old edgeSeeds mechanism was removed.
metadata:
  node_type: memory
  type: feedback
  originSessionId: 4d5176b2-1c69-4013-9932-e601ea592a86
---

Ring topologies need to prime the feedback wire's destination port once at startup, or both sides wait on each other and never start. The current mechanism is a dedicated **bootstrap Input node**: kind `Input`, `data.init=[<seed>]`, `data.repeat=false`, wired by a real edge into the receiving port. It single-fires the seed value at tick-0 and then goes inert.

**Migration history:**
- Pre-RF: edge-level `wire.data.seed` — silently dropped, do not use.
- commit 803873c3 (2026-05-19): migrated to node `initialSlots` (a Go-side loader mechanism).
- commit 8a8dd842 (2026-05-19): replaced by bootstrap Input nodes wired via real edges.
- commit d87cb332 (2026-05-22): inert Go edgeSeeds path deleted.
- branch task/architecture-audit: TS edgeSeeds fossil removed from types, parsers, serializers, and `requiredInputDiagnostics`.

**How to apply:** When wiring a ring, identify the feedback wire (the one closing the loop). Add an Input node whose `topology/nodes/<id>/data.json` is `{"init":[<value>],"repeat":false}` and wire its output to the receiving node's feedback input port via that node's `edges/<label>.json`. The prime is one-shot — it does not re-fire. Do NOT use `data.edgeSeeds` on a node or `data.seed` on an edge — both are removed.

**There is no `topologies/*.json`.** This entry used to say to edit one. A scene is a DIRECTORY TREE (`topology/`, `topology-pair/`) with one directory per node — see `.claude/rules/persistence-ownership.md` for the layout.

**Do not conclude from grep that this entry is stale.** `edgeSeeds` IS still a live identifier in Go — `src/runtopology/streamwire/stream_wiring.go` and `src/Scene/rowtables/row_tables.go` use it for EDGE GEOMETRY seeds (`geomseeds.EdgeGeomSeed`), an unrelated concept that happens to share the word. What was removed is the `data.edgeSeeds` SPEC FIELD, which no longer exists anywhere.
