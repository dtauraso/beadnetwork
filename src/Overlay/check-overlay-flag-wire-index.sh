#!/usr/bin/env bash

# PLACEMENT: src/Input/Codec/messages.ts,src/Input/Codec/input_fingerprint.go | the overlay/panel flag REGISTRY and the wire ordering in INPUT_LAYOUT_FINGERPRINT must list the same flags in the same order

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "$REPO_ROOT"

python3 - <<'PY'
import re, sys

REGISTRY = "src/Input/Codec/messages.ts"
FINGERPRINT = "src/Input/Codec/input_fingerprint.go"

def registry_list(fence):
    src = open(REGISTRY, encoding="utf-8").read()
    m = re.search(fence + r"_START\b(.*?)" + fence + r"_END\b", src, re.S)
    if not m:
        print(f"check-overlay-flag-wire-index: MISCONFIGURED — {fence}_START/_END fence not found in "
              f"{REGISTRY}; refusing vacuous pass", file=sys.stderr)
        sys.exit(1)
    names = re.findall(r'"([A-Za-z0-9_]+)"', m.group(1))
    if not names:
        print(f"check-overlay-flag-wire-index: MISCONFIGURED — parsed 0 names out of the {fence} fence; "
              f"the format changed and this guard would check nothing", file=sys.stderr)
        sys.exit(1)
    return names

def wire_list(marker):
    src = open(FINGERPRINT, encoding="utf-8").read()
    m = re.search(r'InputLayoutFingerprint = "([^"]*)"', src)
    if not m:
        print(f"check-overlay-flag-wire-index: MISCONFIGURED — InputLayoutFingerprint constant not found "
              f"in {FINGERPRINT}; refusing vacuous pass", file=sys.stderr)
        sys.exit(1)
    token = next((t for t in m.group(1).split(" ") if t.startswith(marker)), None)
    if token is None:
        print(f"check-overlay-flag-wire-index: MISCONFIGURED — no {marker} token in the fingerprint; "
              f"refusing vacuous pass", file=sys.stderr)
        sys.exit(1)
    return token[len(marker):].split(",")

fail = False
for fence, marker, what in (
    ("OVERLAY_FLAGS", "overlayFlags=", "overlay"),
    ("PANEL_FLAGS", "panelFlags=", "panel"),
):
    declared = registry_list(fence)
    onwire = wire_list(marker)
    if declared == onwire:
        continue
    fail = True
    print(f"{what.upper()} FLAG WIRE ORDERING DISAGREES WITH THE REGISTRY:")
    print(f"  registry ({REGISTRY}, {fence} fence): {','.join(declared)}")
    print(f"  wire ({FINGERPRINT}, {marker}):       {','.join(onwire)}")
    for name in declared:
        if name not in onwire:
            print(f"  '{name}' is declared but has NO WIRE INDEX — its toggle encodes an index Go cannot")
            print(f"  decode, so the click is dropped at parse and the affordance looks dead.")
    for name in onwire:
        if name not in declared:
            print(f"  '{name}' has a wire index but is not declared — a stale entry shifts the index of")
            print(f"  every flag after it, so toggles land on the WRONG flag.")
    if sorted(declared) == sorted(onwire):
        print(f"  Both lists hold the same names in a DIFFERENT ORDER — the index is positional, so every")
        print(f"  toggle past the first difference addresses the wrong flag.")
    print(f"  Fix: make the {marker} list in {FINGERPRINT} match the {fence} fence exactly, then")
    print(f"  regenerate (go run ./cmd/gen-node-defs).")
    print(f"  Note: check-input-layout-parity stays GREEN through this. input-layout-gen.ts is")
    print(f"  generated FROM the fingerprint, so those two agree by construction whatever the")
    print(f"  registry says — this guard is the only one comparing the wire order to the registry.")

if fail:
    sys.exit(1)
print("check-overlay-flag-wire-index: clean (registry and wire ordering agree for overlay and panel flags)")
PY
