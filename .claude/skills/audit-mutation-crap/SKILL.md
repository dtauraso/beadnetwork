---
name: audit-mutation-crap
description: SLOW-LANE, periodic test-strength audit (NOT per-commit, NOT in stop-checks). Ranks Go functions by CRAP score (crap4go), then runs mutation testing (mutate4go) on the worst-ranked file to check whether surviving tests would notice a broken line.
---

**Slow lane. Do not fold this into `scripts/stop-checks.sh`.** stop-checks is ~3s by
design; both tools here recompile the module and re-run tests per mutation site, so a
single file's mutation run can take minutes. Run this by hand, occasionally — when
`docs/drift-checklist.md` or a real "did we lose test coverage" question comes up, not
on every commit.

## Why this exists

`docs/testing-shape.md` argues cross-goroutine correctness is guaranteed BY
CONSTRUCTION (ownership + message-passing, no locks/atomics), and a prior audit pass
used that argument to delete ~27 tests (see that doc's "History" section). The argument
may well be right, but nothing in this repo currently produces EVIDENCE for or against
it. `tools/check-test-integrity.sh` (if present) only catches tests being WEAKENED over
time — it says nothing about whether the surviving tests had teeth to begin with.

Two tools, two different questions:

- **crap4go** — `CRAP(fn) = CC^2 * (1-cov)^3 + CC`. Says WHERE coverage was lost:
  ranks functions by (complexity x undertested-ness), worst first.
- **mutate4go** — mutation testing. Says whether the tests that DO cover a function
  would actually catch a broken line: it flips an operator (`+`->`-`, `==`->`!=`,
  `&&`<->`||`, etc.), reruns tests, and reports killed / survived / uncovered per site.

crap4go tells you WHERE to spend mutate4go's much higher per-line cost.

## Local tool clones

Both tools are local clones, each its own Go module (not vendored into this repo):

```
~/Downloads/unclebob-repos/crap4go
~/Downloads/unclebob-repos/mutate4go
```

Read each clone's own README (and crap4go's bundled `SKILL.md`, mutate4go's bundled
skill under its own `skills/` directory) for the authoritative flag list before inventing
invocations — this doc summarizes, it does not replace them. Use
`tools/audit-mutation-crap.sh`, which builds each tool to a temp binary (they're
separate modules, so `go run <path>` from inside this repo fails with "outside main
module").

## Workflow

1. **Rank.** From the repo root:

   ```bash
   tools/audit-mutation-crap.sh crap nodes
   ```

   This runs `go test ./... -coverprofile=target/coverage/coverage.out`, then prints a
   table of every function sorted by CRAP score, worst first. Filter arguments
   (`nodes`, `nodes/Wiring`) narrow which functions are REPORTED — the coverage run
   itself still covers the whole module unless you also pass `--test-command`.

   Confirmed working 2026-07-28: on this repo's `nodes/` the worst-ranked functions
   were `nodeMover.writeStreamFrame` (CC 17, 4.0% cov, CRAP 272.7) and
   `edgeMover.writeStreamFrame` (CC 15, 4.9% cov, CRAP 208.7), both in
   `nodes/Wiring/port_geom_emit.go` — the geometry-streaming path, which makes sense as
   an audit target: high branch count, low coverage, and it's the kind of thing
   `docs/testing-shape.md`'s corollary 2 would say not to unit-test directly (it's a
   goroutine's own emission, which IS a legitimate shape per that doc — so mutate4go
   applies here, this isn't a cross-goroutine test).

2. **Scan one file (cheap, safe).** Before running a real mutation pass, check the site
   count with the read-only structural mode:

   ```bash
   tools/audit-mutation-crap.sh mutate-scan nodes/Wiring/port_geom_emit.go
   ```

   `--scan` skips coverage and test execution entirely and writes NOTHING to the
   source file — confirmed by reading mutate4go's manifest package
   (`internal/manifest` in the mutate4go clone): its `Embed()`/`os.WriteFile` are only
   reached from the non-scan path.

3. **Mutate ONE file at a time.** mutate4go recompiles and reruns tests per mutation
   site, so it does not scale past a single file per invocation:

   ```bash
   tools/audit-mutation-crap.sh mutate nodes/Wiring/port_geom_emit.go --max-workers 3
   ```

   Loop per the upstream README's "Recommended Workflow": if a mutation is UNCOVERED,
   add/fix a test until it's covered; if a mutation SURVIVES, the surviving assertion is
   too weak — strengthen the test or fix the code; rerun the same file until clean
   before moving to the next-worst file from the crap4go ranking.

## THE TRAP: mutate4go writes into your source file

A real (non-`--scan`) mutate4go run — including `--update-manifest` — embeds a footer
comment directly into the target `.go` file on success:

```go
// mutate4go-manifest-begin
// {"version":1,"tested_at":"...","module_hash":"...","functions":[...]}
// mutate4go-manifest-end
```

(in the mutate4go clone's manifest package: `Build()` computes it, `Embed()` appends
it, the CLI writes it with `os.WriteFile(path, ...)`.) It exists to make the tool's
DIFFERENTIAL mode fast on a second run (only re-mutates functions whose hash changed),
but a stray committed manifest — or a leftover `<file>.mutate4go.bak` backup written
next to the source mid-run — will trip THREE of this repo's stop-checks guards:

- `tools/check-gofmt.sh` — the rewritten file may not be gofmt-clean.
- `tools/check-no-untracked-source.sh` — a surviving `.mutate4go.bak` is untracked
  source this guard is specifically built to catch.
- `tools/check-comment-vocab.sh` — any banned token the manifest JSON happens to echo
  (unlikely but not impossible: function names land verbatim in the JSON) would trip
  the retired-vocabulary scan.

**Safe invocation:**

- Use `mutate-scan` / `--scan` for anything you are not about to act on immediately —
  it never writes.
- After a real run, before committing anything: `git diff <file>` the target file. If
  you don't deliberately want differential mode going forward, `git checkout --
  <file>` to discard the manifest and start clean next time; a full non-differential
  rerun is cheap to redo later (`--mutate-all`).
- Always finish with `git status --porcelain` and delete any `*.mutate4go.bak`
  stragglers (a crashed worker can leave one) before committing anything.

## Coverage artifact

Both tools write `target/coverage/coverage.out`, matched by this repo's `*.out`
`.gitignore` entry. An untracked `target/coverage/coverage.out` was already sitting in
the tree before this skill was written — evidence of a prior ad hoc run, not a problem
by itself since it's ignored, but confirm with `git status --porcelain -uall` after any
run here that nothing under `target/` needs staging.
