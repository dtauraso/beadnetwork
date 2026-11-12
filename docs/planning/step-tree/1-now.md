---
branch: task/step-tree
---

# 1 · Steps on disk now

Taken from `topology-pair-phi-theta-3/sequences/8/` and the Builder code that writes it
(`Categories/Chrome/Pills/Builder/step_*.go`).

```
sequences/8/
  name.bin                      "sequence 8"
  view/s.bin
  nodes/                        the network: what runs
    2/drag/ref-*.bin
    3/drag/links/2To3/          start-*, delta-*, increment-*, one-*, lock-*, picked, source
  step/
    lock-*.bin                  network-wide step state   (m.Step)
    1/ 2/ 3/                    per-NODE step state       (m.Steps, keyed by node id)
    heads/step-1/lock-*.bin     step 1's head row
    heads/step-1/2/lock-*.bin   step 1, node 2's head
    picks/step-1/2To3.bin       step 1 picks link 2→3
    clones/step-1/3/links/2To3/ step 1's copy of link 2→3
    clones/step-2/...
```

## What is wrong with it

| # | Problem | Evidence |
|---|---|---|
| 1 | One step is split across three trees. To see step 1 you open `heads/step-1`, `picks/step-1` and `clones/step-1`. | `step_heads.go:21`, `step_picks.go:60`, `step_links.go:16` |
| 2 | `step/1/` is node 1, but `heads/step-1/` is step 1. The same name means two different things in the same folder. | `step_files.go:108` writes `m.Steps[nodeID]` to `step/<id>` |
| 3 | The list of steps lives in code, not on disk: `var StepNumbers = []int{1, 2}`. A third step needs a code change. | `step_picks.go:19` |
| 4 | A pick is stored twice: `picks/step-1/2To3.bin` and `picked.bin` inside the clone's link directory. | dumped files |
| 5 | Folder names are `step-1` here, while sequences are `1/`, `2/`. Two naming rules for one kind of thing. | `Scenes/sequences.go:35`: "sequences/ is an array indexed by its directory names" |
