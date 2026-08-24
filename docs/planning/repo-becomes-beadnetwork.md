# Repo becomes beadnetwork

## Target

`wirefold` names two things the model does not have. `MODEL.md` and
`docs/model/entities.md` never say "wire", and nothing folds. The entities are
bead, bead line, node goroutine, node input, clock. `beadnetwork` names the
whole at the right scale — a network of nodes, with beads moving through it —
and both halves already mean exactly that here.

## Ripple list

Five kinds of name, in increasing order of who else notices.

1. **Go module path.** `go.mod`'s `module github.com/dtauraso/wirefold`, and the
   import line in ~300 `.go` files. Compiler-checked: miss one and the build
   fails by name.
2. **npm package and bundle.** `package.json` `name`, `esbuild.mjs`, `tsconfig`.
   Also compiler/bundler-checked.
3. **VS Code contribution ids.** `wirefold.probe.trace`,
   `wirefold.reloadOnHostBuild`, and the command ids. NOT checked by anything:
   a stale id silently reads as unset, which is the failure mode to watch.
4. **Environment variables.** Twelve `WIREFOLD_*` names across 52 files, read by
   Go and written by the extension host. A missed pair fails silently the same
   way — Go reads empty and degrades rather than erroring.
5. **Prose.** 20 markdown/shell files, plus guard `PLACEMENT:` headers.

Deliberately NOT in this change, because both break the running session and are
the user's call:

- the checkout directory name on disk
- the GitHub remote (`gh repo rename`)

## Order

Module path first, since the compiler proves it. Then package/bundle. Then the
ids and env vars, which nothing proves — so each is renamed by exhaustive grep
for the literal string, and the grep is re-run to zero afterwards. Prose last.

## Verification

- `go build ./...` and `stop-checks.sh` prove 1, 2 and 5.
- 3 and 4 are proven by grep-to-zero on `wirefold` and `WIREFOLD_`, case
  insensitive, across the tree.
- The settings/env pair is proven at RUNTIME, not by grep: run the binary with
  `BEADNETWORK_PROBE_TRACE=1` and confirm a trace file appears, since that is
  the exact link a rename can sever without any check firing.
- Byte comparison of the block files against `main`, 3 runs each.

## Risks

- **A live install keeps the old setting.** Anyone with `wirefold.probe.trace`
  set in their VS Code settings loses it; the new key reads as unset (default
  off). Nothing warns. This is the one user-visible consequence.
- **Grep-to-zero is weaker than it looks** for env vars: a name assembled from
  parts would not match. Checked by reading the reader and the writer of each
  variable, not just the count.
