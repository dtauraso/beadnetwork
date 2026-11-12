---
branch: task/step-tree
---

# 2 · Proposed tree: a step is a directory

A step and everything about it share one directory, which is the repo's own layout rule.
Inside, a step uses the same shape as the network (`nodes/<id>/links/<from>To<to>/`), so the
reader that loads the network's links (`Links.ReadLinks`) also loads a step's links.

```
sequences/8/
  name.bin
  view/s.bin
  nodes/                          the network (unchanged)
    3/drag/links/2To3/
  steps/                          an array, indexed by directory name, like sequences/
    1/
      name.bin                    optional label shown on the step card ("lift 3")
      lock-*.bin                  the step's head row          (was step/heads/step-1/*.bin)
      nodes/
        2/lock-*.bin              node 2's head in this step   (was step/heads/step-1/2/)
        3/links/2To3/             this step's copy of 2→3      (was step/clones/step-1/3/links/2To3/)
          picked.bin              the pick, kept in one place  (was step/picks/step-1/2To3.bin)
          start-*.bin delta-*.bin increment-*.bin one-*.bin
    2/
      ...
```

## Rules

1. **The steps are the directories under `steps/`.** `StepNumbers` is deleted. Reading the
   directory gives the list, sorted numerically, and a name that is not a number panics with a
   site tag, the same way `Scenes.Sequences` does.
2. **A step holds only what it differs on.** If a link has no directory under the step, the
   step uses the network's value. The current seeding (`SeedStepLinks` copies every link into
   every step) is kept for now, so this is a later choice. See page 5.
3. **There is one copy of each fact.** A pick is `picked.bin` in the step's link directory and
   nowhere else.
4. **The network's own step state moves out of the way**, so `step/` no longer exists:
   `step/lock-*.bin` → `press/lock-*.bin`, and `step/<id>/` → `nodes/<id>/press/`.
   ("press" because that state is what the step press buttons use.) This name is still open.

## Grouping sequences (optional, same idea one level up)

```
sequences/
  1/  2/  ...                     a directory with name.bin is a sequence
  lifts/                          a directory without name.bin is a folder
    1/  2/
```

The sequence list then reads like a file browser, with folders that fold. This is only worth
doing once there are enough sequences to need it.
