---
name: project_node_color_vocab
description: David's shorthand for node kinds — "time nodes" = Time (renamed from HoldNewSendOld), "and nodes" = WindowAndInhibit*Gate
metadata:
  type: project
---

David refers to some node kinds by a nickname. Mapping to the Go/`NODE_DEFS` kinds:

- **"time nodes"** = `Time` (kind + struct renamed from `HoldNewSendOld` on task/rename-holdnewsendold-time; the nickname now matches the kind name).
- **"and nodes"** = `WindowAndInhibitLeftGate` / `WindowAndInhibitRightGate` (the AND gates).

Use these terms when he does. Other kinds seen in the sample topology: `Input`, `Pulse`,
`Pacer`, `Hold` — no nickname given yet.
