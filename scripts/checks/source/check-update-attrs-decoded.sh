#!/usr/bin/env bash
set -euo pipefail

# PLACEMENT: **/update_attrs.go | every attr a concern declares must be bound to a wire byte and compared by that concern's decoder, or an edit naming it decodes to nothing

# Replaces check-input-attr-dispatched.sh, which compared a decoder in one package
# against a dispatch table in another. Both halves now live in the concern that owns
# the attr, so the drift it caught is a drift WITHIN one package: a name in
# UpdateAttrs that no decoder ever compares. Such an edit crosses the wire intact,
# decodes to nothing, and the affordance looks dead with nothing reporting why.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "$REPO_ROOT"

python3 - <<'PY'
import os, re, sys, glob

files = [p for p in glob.glob("Categories/**/update_attrs.go", recursive=True)]
if not files:
    print("check-update-attrs-decoded: MISCONFIGURED — no */update_attrs.go exists.")
    print("Per-concern attr declarations were renamed or removed; this guard is scanning")
    print("nothing rather than passing. Repoint it.")
    sys.exit(1)

fail = 0
checked = 0
for decl in sorted(files):
    pkgdir = os.path.dirname(decl)
    src = open(decl).read()
    m = re.search(r"var UpdateAttrs = \[\]string\{(.*?)\}", src, re.S)
    if not m:
        print(f"check-update-attrs-decoded: MISCONFIGURED — {decl} declares no")
        print("'var UpdateAttrs = []string{...}'. The declaration's shape changed.")
        sys.exit(1)
    names = re.findall(r'"([^"]+)"', m.group(1))
    if not names:
        print(f"check-update-attrs-decoded: MISCONFIGURED — {decl} has an empty UpdateAttrs.")
        sys.exit(1)

    others = [p for p in glob.glob(os.path.join(pkgdir, "*.go"))
              if p != decl and not p.endswith("_test.go")]
    body = "\n".join(open(p).read() for p in others)

    for name in names:
        checked += 1
        bind = re.search(r"(\w+)\s*=\s*attrIndex\(" + re.escape(f'"{name}"') + r"\)", body)
        if not bind:
            print(f'{pkgdir}: attr "{name}" is declared in UpdateAttrs but no attrIndex("{name}")')
            print("  binds it to a wire byte. TS can encode an edit naming it; Go has no constant")
            print("  to compare, so the edit decodes to nothing and the affordance looks dead.")
            fail += 1
            continue
        ident = bind.group(1)
        uses = len(re.findall(r"\b" + re.escape(ident) + r"\b", body)) - 1
        if uses < 1:
            print(f'{pkgdir}: attr "{name}" binds {ident}, but {ident} is never compared against')
            print("  the incoming attr byte. The edit arrives, matches no case, and is dropped")
            print("  silently — the click does nothing and nothing reports why.")
            fail += 1

if fail:
    print()
    print(f"check-update-attrs-decoded: {fail} undecoded attribute(s) across {len(files)} concerns")
    sys.exit(1)
print(f"check-update-attrs-decoded: clean ({checked} attrs across {len(files)} concerns)")
PY
