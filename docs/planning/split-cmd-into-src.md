# The generators live with the thing each one generates, and cmd/ is removed

## The target

`cmd/` holds one binary, `cmd/gen-node-defs`: 9 subpackages plus 720 lines at its root,
writing 30 generated files across `src/`, `nodes/` and the repo root. It is the last
directory that groups code by what it *is* (an executable) rather than by what it is
*about*, which is the layout rule the rest of the repo already follows — "a thing and
everything about it share a directory".

After this change there is no `cmd/`. Each generator sits in the concern directory whose
file it generates, as that concern's own `main` package, and `go generate ./...` runs them
all.

## Why

`cmd/gen-node-defs/buflayout` is the buffer layout. `src/schema/buffer-layout/` is the
buffer layout. They are the same subject in two places, and the only thing separating them
is that one half compiles to an executable. The same split exists for wire defs, trace
kinds, overlay tables, input layout and scenes. Every generated file already carries a
header naming `cmd/gen-node-defs`, which is a pointer that exists only because the source
is somewhere the reader would not look.

## The Go constraint that shapes this

**A directory is one Go package.** `src/Buffer/` already holds library Go, so a `main`
package cannot join it. Each generator therefore lands in a `gen/` subdirectory of its
concern — `src/Buffer/gen/`, `src/schema/buffer-layout/gen/` — and the concern's existing
Go file carries the `//go:generate` line that runs it. This is what "beside" can mean in
Go; it is one directory further out than the prose suggests, and it is not optional.

## Destination table

| moves | to | writes |
|---|---|---|
| `buflayout/` (minus `column_streams.go`) | `src/schema/buffer-layout/gen/` | `buffer-layout*.ts`, `src/Buffer/buffer_layout_gen*.go` |
| `params/` + `constexpr/` | `src/schema/buffer-layout/gen/` | `curve-params.ts`, `shading-params.ts`, `nodes/Wiring/nodegeom/{curve,shading}_params.go` |
| `buflayout/column_streams.go`, `gen_pipelines_columns.go` | `src/Buffer/gen/` | every `*/columns-gen.ts`, `column_streams_gen.go` |
| `frame_tags.go` | `src/Buffer/gen/` | `frame_tags.go`, `frame-tags.ts` |
| `inputlayout/` | `src/schema/input/gen/` | `input-layout-gen.ts` |
| `nodedefs/` | `src/schema/gen/` | `node-defs.ts`, `nodes/Wiring/loadspec/topo_spec.go` |
| `wiredefs/` | `src/schema/gen/` | `wire-defs.ts` |
| `gen_scenes.go` | `src/schema/gen/` | `scenes-gen.ts`, `nodes/Wiring/scene/scene.go` |
| `tracekinds/` | `src/Trace/gen/` | `trace-kinds.ts` |
| `overlaygen/` | `src/OverlaysDropdown/gen/` | `overlay_state.go`, `overlay_tables_gen.go` |
| `node_dims.go` | `nodes/Wiring/nodegeom/gen/` | `node_dims_gen.go`, `src/Node/node_kind_id_gen.go` |
| `kindscan/` | `nodes/kindscan/` | nothing — it *reads* `nodes/<Kind>/*.go` and `SPEC.md` |
| `main.go`, `repo_root.go`, `src_root.go` | deleted / inlined | — |

Several generators write into two trees at once (a Go file under `nodes/` and a TS file
under `src/`). The home is the concern the generated thing *belongs* to, not whichever
path is written first: curve params ride in the buffer, so they live with the buffer layout
even though one output lands in `nodegeom`.

**`kindscan` and `constexpr` are the two that generate nothing.** `constexpr` has exactly
one consumer (`params/params_shading.go`) and moves in with it. `kindscan` has five, and
its subject is the node kinds themselves — it parses `nodes/<Kind>/*.go` ASTs and
`SPEC.md`. It moves to `nodes/kindscan/`, beside what it reads, rather than being split by
consumer: splitting it would duplicate the AST walk in five places. If that reading of
"split them too" is wrong, this row is the one to change, and it is worth changing before
step 2 rather than after.

## Ripple list

- **`main()` disappears and its ordering with it.** The current `main.go` calls fourteen
  generators in a fixed sequence. Check first whether any later call depends on an earlier
  one's output; `go generate ./...` walks packages in directory order and gives no such
  guarantee. If a dependency exists, it must become an explicit input, not an ordering.
- **`resolveRepoRootAndKinds`** collects and ID-assigns kinds once, then hands the slice to
  five generators. Split apart, each binary redoes it. That is fine (it is a filesystem
  walk, not state), but `AssignKindIDs` must be deterministic across processes or node kind
  IDs will differ per generator — verify before splitting, not after.
- **`repo_root.go` / `src_root.go`** are found by walking up for `nodes/` and by locating
  the single `package.json`. Every binary needs both. They are 60 lines total; duplicating
  them into each `gen/` is worse than a small shared package.
- **`package.json:49`** — `"gen:node-defs": "go run ./cmd/gen-node-defs"` becomes
  `go generate ./...`, and the script name stops being accurate.
- **`main.go:3`** carries the repo's only `//go:generate` line. It is replaced by one line
  per concern.
- **`scripts/checks/meta/check-dir-size.sh:16,18,39`** and
  **`check-file-size.sh:28`** walk `cmd/gen-node-defs` by literal path. They must gain the
  new directories, or the size ceilings silently stop applying — a guard walking a path
  that no longer exists still exits 0.
- **`nodes/check-spec-format-view-fields.sh:10`** hardcodes `GEN_DIR="cmd/gen-node-defs"`.
- **`src/check-generated.sh:3`** — its `PLACEMENT:` header names the old command.
- **30 generated files** carry a header comment naming `cmd/gen-node-defs`. These are
  regenerated, so they change by rerunning, not by editing — but the header text is emitted
  by the generators and must be updated in the generator source first.
- **`ARCHITECTURE.md`, `CLAUDE.md`, `.claude/rules/node-kinds.md`,
  `.claude/rules/buffer-schema.md`, `nodes/SPEC-FORMAT.md`, `nodes/PairNode/SPEC.md`,
  `nodes/PairNode/docs/BEHAVIOR.md`** all name the command or the directory. CLAUDE.md
  states "**`cmd/`** — the generators, where Go keeps executables" as repo doctrine and the
  primitive landing rule names `go run ./cmd/gen-node-defs` as step 4; both are reversed by
  this change.
- **`check-guard-paths-exist.sh`** should catch any path reference missed above. It is the
  backstop, not the survey.

## Order

1. Settle the `kindscan` row.
2. Move `kindscan` and `constexpr` first — they have no outputs, so nothing regenerates and
   the move is pure import rewriting.
3. Move one generator end to end (`tracekinds` is smallest: 2 files, 1 output) and confirm
   `go generate ./...` reaches it and reproduces `trace-kinds.ts` byte for byte.
4. Move the rest, one generator per commit, regenerating after each.
5. Delete `cmd/`, update `package.json`, and update the guards that name the old path.
6. Update the prose pages last, once the layout is settled.

## Verification

Byte-for-byte reproduction is the check: run the current `go run ./cmd/gen-node-defs` on
`origin/main`, keep the 30 outputs, then after the split run `go generate ./...` and diff.
Any difference is a bug in the move, except header comments, which change deliberately and
should be diffed separately.

Then `bash scripts/stop-checks.sh` with EMPTY stdout — it runs `check-generated.sh`,
`check-dir-size.sh`, `check-file-size.sh` and `check-guard-paths-exist.sh`, which together
cover most of the ripple list.

## Risks

- **The silent-guard class.** Three guards walk `cmd/gen-node-defs` by literal path. When
  it stops existing they find zero files and pass. `check-guard-paths-exist.sh` exists for
  exactly this and must be confirmed to fire before it is trusted — break one path on
  purpose and watch it fail by name.
- **One command becomes fourteen.** The landing rule's force comes from "skip
  `go run ./cmd/gen-node-defs` and the kind does not exist, while everything else looks
  correct". `go generate ./...` must be verified to reach every new `gen/` package, or that
  failure mode gets quieter rather than louder.
- **Kind-ID drift.** See the ripple list: five binaries independently assigning node kind
  IDs is a wrong-by-construction hazard if `AssignKindIDs` is order- or count-sensitive.
- **`gen/` reads as scaffolding.** The directory name is imposed by Go's one-package rule,
  not chosen. If a concern's `gen/` later grows something that is not a generator, the name
  stops being true.
