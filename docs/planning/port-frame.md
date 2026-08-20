# Remove the port frame: one declaration, generated

## The target

A node kind declares its named inputs and outputs ONCE, in `SPEC.md`'s `## Ports` table.
Go's per-kind `[]portwiring.PortSpec` argument to `RegisterBuilder` is GENERATED from that
table, not hand-written, so the word "port" names only generated internals and nothing a
human maintains twice.

## Why now

Today a kind declares the same thing twice: the SPEC.md table (which generates
`node-defs.ts`, what the editor draws) and the Go list (what the runtime binds channels to).
The check that tied them together read the kind's Go struct fields through the AST. That AST
path was deleted, and the check left behind compares `parsePortsFromSpec` against
`parseSpecMD` — both sides are SPEC.md, so it cannot fail. Its message still says "does not
match any Go channel-typed port", naming a check that is gone.

Demonstrated, not assumed: renaming `FeedbackIn` to `FeedbackTypo` in `input/SPEC.md`
propagates into `node-defs.ts` with no complaint while `input/node.go` still registers
`FeedbackIn`. The editor and the runtime disagree about a port name and nothing notices.

## The invariant this reverses

`.claude/rules/node-kinds.md` currently says the kind "declares its ports as an explicit
`[]portwiring.PortSpec` argument … rather than having them reflected off its struct, which is
why a forgotten field is now a compile error instead of a silently nil one." That reasoning
was against REFLECTION, and it still holds — generation is not reflection: the names come from
the authored table, and a kind whose table is missing fails loudly at registration rather than
binding a nil channel.

## Ripple list

1. `src/Node/Wiring/nodegeom/gen` — third output, `portwiring/kind_ports_gen.go`: a
   `KindPorts map[string][]PortSpec` keyed by Go kind name. Direction mapping: `in` → `PortIn`,
   `out` → `PortOut`, `broadcast` → `PortBroadcast`.
2. `kindapi.RegisterBuilder` — loses the `ports` parameter; reads `portwiring.KindPorts[kind]`
   and panics naming the kind when the table is absent.
3. The 15 `src/NodeKinds/<Kind>/` registrations — drop the port literal and the
   `portwiring` import where it becomes unused.
4. `kindscan/scan.go` — delete the self-comparing check and the `specPortNames` return it
   consumed, if nothing else uses it.
5. Docs: `.claude/rules/node-kinds.md` (the passage above), `SPEC-FORMAT.md` if it claims a Go
   declaration, CLAUDE.md's node-kind landing rule if it names the argument.

## Order

Generator first, so `KindPorts` exists before anything reads it; then `RegisterBuilder`; then
the kinds, which is the step that must not compile until both are in place. The vacuous check
goes last, once nothing depends on its return.

## Verification

`go generate ./...` must leave `node-defs.ts` byte-identical — the table is the source in both
paths, so the editor's view of ports cannot change. Then the drift test that motivated this:
rename a port in one SPEC.md and confirm the build now FAILS by name instead of generating a
disagreement. Restore, and `bash scripts/stop-checks.sh` empty.

## Risk

`RegisterBuilder` runs in `init()`, so a missing table is a panic at process start, not a
compile error — louder than the silent drift it replaces, but later. The generated file must
therefore be committed, and `check-generated.sh` is what keeps it current.
