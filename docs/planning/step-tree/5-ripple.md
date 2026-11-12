---
branch: task/step-tree
---

# 5 · Ripple, order, verification, open questions

## Code that changes

| File | Change |
|---|---|
| `Builder/step_picks.go` | delete `StepNumbers`, add `StepNumbers(stepsDir)` that reads the directory; picks are read from and written to `picked.bin` |
| `Builder/step_links.go` | `stepClonesDir` → `steps/N/nodes`; `DirStepClones` is deleted |
| `Builder/step_heads.go` | `stepHeadsDir` → `steps/N`; `DirStepHeads` is deleted |
| `Builder/step_files.go` | `DirStep` → `press`; `m.Steps[id]` is written to `nodes/<id>/press/` |
| `Builder/pill_links.go` | loop over the step list read from disk |
| `Scene/Startup/fit_s.go:66` | read starts from `steps/*/nodes` |
| `Scene/Dispatch/sheet_hits.go` | new ops for add, duplicate, delete and move step (only if page 4's buttons are built) |
| new `scripts/migrate-steps/` | one-time migration over every scene |

## Order

1. The path move alone: the migration script plus the five Builder and Startup files. Nothing
   changes in what you see, so if the editor looks different, the migration is wrong.
2. Remove the duplicate pick store (`picks/`).
3. The step list comes from disk (`StepNumbers` is deleted).
4. The add, duplicate and delete buttons, if you want them.

## Verification

- `bash scripts/stop-checks.sh` prints nothing.
- Open `topology-pair-phi-theta-3` sequence 8 before and after step 1. Every card shows the
  same numbers, and 2→3 draws on the same tick.
- `find */sequences/*/step` finds nothing after the migration.

## Open questions

1. The name for the network's own step state: `press/`, or something else?
2. Should a step hold only what it differs on (rule 2 on page 2), or keep today's full copy of
   every link? A sparse step is smaller, but a later network edit would then show up in every
   step that does not override that link.
3. Folders for sequences: now, or later?
