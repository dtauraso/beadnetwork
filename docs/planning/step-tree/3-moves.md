---
branch: task/step-tree
---

# 3 · Old path → new path, and step operations as file operations

## Path map (for a one-time migration of existing scenes)

| Old | New |
|---|---|
| `step/clones/step-N/<to>/links/<f>To<t>/` | `steps/N/nodes/<to>/links/<f>To<t>/` |
| `step/picks/step-N/<f>To<t>.bin` | `steps/N/nodes/<t>/links/<f>To<t>/picked.bin` |
| `step/heads/step-N/lock-*.bin` | `steps/N/lock-*.bin` |
| `step/heads/step-N/<id>/lock-*.bin` | `steps/N/nodes/<id>/lock-*.bin` |
| `step/lock-*.bin` | `press/lock-*.bin` |
| `step/<id>/lock-*.bin` | `nodes/<id>/press/lock-*.bin` |

The migration is a script run once over every `*/sequences/*/`, after which the old paths are
gone. The code never reads both layouts.

## What each step action does on disk

| Action in the editor | On disk |
|---|---|
| Add a step after step N | copy `steps/N/` to `steps/N+1/`, after shifting the later steps up by one |
| Duplicate step N | same as add, so the new step starts as an exact copy |
| Delete step N | remove `steps/N/`, then shift the later steps down by one |
| Move step N before step M | rename the directories in order |
| Name a step | write `steps/N/name.bin` |
| Edit a link in step N | write `steps/N/nodes/<to>/links/<f>To<t>/<field>.bin` (same as now, new path) |
| Pick a link in step N | write `picked.bin` in that link's directory |

Every action is a directory operation, so you can also do it in Finder or a terminal while the
editor is closed, and the editor reads the result on the next open.
