#!/usr/bin/env bash

# PLACEMENT: .vscode/launch.json | F5 is the only way into the Extension Development Host; a launch.json that lost its config leaves no way to run the editor, and no compiler reads JSON

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "$REPO_ROOT"

python3 - <<'PY'
import json
import pathlib
import sys

LAUNCH = pathlib.Path(".vscode/launch.json")
WANT = "Run Extension"

if not LAUNCH.exists():
    print(f"check-launch-config: {LAUNCH} not found — F5 has nothing to launch, so there is "
          f"no way to open the Extension Development Host.", file=sys.stderr)
    sys.exit(1)

raw = LAUNCH.read_text(encoding="utf-8")
stripped = "\n".join(l for l in raw.splitlines() if not l.strip().startswith("//"))
try:
    doc = json.loads(stripped)
except json.JSONDecodeError as e:
    print(f"check-launch-config: {LAUNCH} does not parse ({e}) — VS Code would ignore it "
          f"and F5 would do nothing.", file=sys.stderr)
    sys.exit(1)

configs = doc.get("configurations") or []
match = [c for c in configs if c.get("name") == WANT]
if not match:
    names = ", ".join(repr(c.get("name")) for c in configs) or "none"
    print(f"check-launch-config: {LAUNCH} has no configuration named {WANT!r} (found: {names}).")
    print(f"  That entry IS the way into the Extension Development Host. Without it F5 opens")
    print(f"  nothing, the extension never loads, and 'Topology: Open Editor' is absent from")
    print(f"  the palette — which looks like a broken extension rather than a missing config.")
    sys.exit(1)

cfg = match[0]
args = cfg.get("args") or []
dev = [a for a in args if str(a).startswith("--extensionDevelopmentPath=")]
if not dev:
    print(f"check-launch-config: {WANT!r} passes no --extensionDevelopmentPath, so the host "
          f"launches without this extension.", file=sys.stderr)
    sys.exit(1)

path = dev[0].split("=", 1)[1].replace("${workspaceFolder}", ".").strip()
manifest = pathlib.Path(path) / "package.json"
if not manifest.is_file():
    print(f"check-launch-config: --extensionDevelopmentPath points at {path!r}, which holds no "
          f"package.json — the host would find no extension manifest there.")
    print(f"  The package root moves; this argument has to move with it.")
    sys.exit(1)

folder_args = [a for a in args if not str(a).startswith("--")]
if not folder_args:
    print(f"check-launch-config: {WANT!r} passes no folder for the host to open. The extension "
          f"resolves the topology tree from the first workspace folder, so a host opened on "
          f"nothing finds nothing.", file=sys.stderr)
    sys.exit(1)

opened = folder_args[-1].replace("${workspaceFolder}", ".").strip()
if not (pathlib.Path(opened) / "go.mod").is_file():
    print(f"check-launch-config: the host is told to open {opened!r}, which holds no go.mod.")
    print(f"  The extension reads workspaceFolders[0] to find the topology tree, so it must be")
    print(f"  handed the repo root, not a subfolder.")
    sys.exit(1)

print(f"check-launch-config: clean ({WANT!r} develops {path!r}, opens {opened!r})")
PY
