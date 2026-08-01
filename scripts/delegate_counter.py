#!/usr/bin/env python3
"""Shared state for the two delegation hooks.

force-delegate-hook.py (PreToolUse) and delegate-reminder-hook.py
(UserPromptSubmit) both address the same per-session counter, so the path rule
and the file format live HERE rather than being restated in each hook. Two
copies of `f"/tmp/claude-delegate-{safe}.count"` is exactly the string
duplication audit-grep-load flags: a rename in one hook would silently give the
other its own private counter, and nothing would fail -- the reset would just
stop working.

`dirty_paths` answers "which files is this session in the middle of changing?"
from git, not from a recorded edit list. A hook only sees the tools its matcher
names, so recording Edit/Write paths would require widening the PreToolUse
matcher in .claude/settings.json; the working tree already knows, needs no
matcher, and survives a hook that never fired. It also counts edits made by the
person driving the editor, which is correct for the same reason: a file with
uncommitted changes is live work, whoever changed it.
"""
import os
import re
import subprocess

BASE = "/tmp/claude-delegate-"


def counter_path(session_id: str) -> str:
    safe = re.sub(r"[^A-Za-z0-9_-]", "_", session_id or "default")
    return f"{BASE}{safe}.count"


def read_count(session_id: str) -> int:
    try:
        with open(counter_path(session_id)) as f:
            return int(f.read().strip() or "0")
    except Exception:
        return 0


def write_count(session_id: str, n: int) -> None:
    try:
        with open(counter_path(session_id), "w") as f:
            f.write(str(n))
    except Exception:
        pass


def reset(session_id: str) -> None:
    write_count(session_id, 0)


def dirty_paths(cwd: str) -> set:
    """Repo-relative paths with uncommitted changes, plus their basenames.

    Basenames are included because the exemption has to recognise a path inside
    a Bash command string, where it may be written relative to any directory.
    A false exemption costs one uncounted lookup; a false BLOCK costs a
    round-trip, so the asymmetry is deliberate. Empty on any failure (not a git
    repo, git missing, timeout) -- the budget then simply applies as before.
    """
    try:
        out = subprocess.run(
            ["git", "status", "--porcelain", "-uall"],
            cwd=cwd or os.getcwd(), capture_output=True, text=True, timeout=5,
        ).stdout
    except Exception:
        return set()
    paths = set()
    for line in out.splitlines():
        p = line[3:].strip()
        if " -> " in p:  # rename: "old -> new"
            p = p.split(" -> ", 1)[1]
        p = p.strip('"')
        if p:
            paths.add(p)
            paths.add(p.rsplit("/", 1)[-1])
    return paths
