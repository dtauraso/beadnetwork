#!/usr/bin/env python3
"""PreToolUse hook: hard-block open-ended executor-style work.

Counts Read / Grep / Glob calls plus Bash calls whose command starts with a
search verb (grep, rg, find, ls, cat, head, tail, awk, sed), and denies past a
threshold. The Task/Agent tool resets the counter.

WHAT THE THRESHOLD IS FOR, and the two things that used to trip it wrongly
-------------------------------------------------------------------------
The cost this hook exists to control is context: an inline lookup dumps whole
file ranges into the main thread, a subagent returns conclusions. But raw call
COUNT is a poor proxy for that cost, and it misfired twice in one session
(2026-08-01) in the same shape -- it fired on work that was neither open-ended
nor exploratory:

  1. Mid-implementation. Three `sed` ranges were needed to fix three compile
     errors the session had just CAUSED and already understood. Delegating that
     made a subagent rediscover held context; it fixed the call sites, and also
     left a stray built binary at the repo root.
  2. Across a task boundary. Lookups spread over three unrelated tasks tripped
     the same wire as nine lookups chasing one question.

So the counter now resets at a TASK BOUNDARY, and re-reading a file this
session already EDITED is free:

  - Every user prompt resets it (delegate-reminder-hook.py) -- the budget is per
    instruction, not per session.
  - Starting or switching a task branch resets it (TASK_BOUNDARY below).
  - Read/sed/grep of a path with UNCOMMITTED CHANGES is exempt entirely -- not
    counted, never blocked. Reading a file you are in the middle of changing is
    the tail of a known edit, not exploration. The live-work set comes from
    `git status` (delegate_counter.dirty_paths) rather than a recorded edit
    list, so it needs no widening of this hook's PreToolUse matcher.

Nine lookups answering ONE open question is still the thing this blocks, and
that is unchanged.

ASK, NOT DENY. A hard deny could collide with a session whose own instructions
forbid spawning subagents unasked: neither rule could be satisfied, and the only
legal move left was to interrupt the user for a ruling -- a whole conversation
turn, spent by the rule that exists to save cost. `ask` puts the same decision
in front of the same person as a one-click permission prompt instead. The
friction survives (it is still visible, still interrupts the reflex); the
deadlock does not, because allow/deny is always a legal answer.
"""
import json
import re
import sys

sys.path.insert(0, __file__.rsplit("/", 1)[0])
from delegate_counter import read_count, write_count, reset, dirty_paths  # noqa: E402

THRESHOLD = 8  # block on the (THRESHOLD+1)th call
SEARCH_VERBS = re.compile(r"^\s*(grep|rg|find|ls|cat|head|tail|awk|sed)\b")

# Bash commands that START a new piece of work. Matching one means the lookups
# that follow belong to a different question than the ones before it, so the
# budget starts over. new-task.sh is this repo's own branch-per-change entry
# point (CLAUDE.md Workflow); the git forms cover the same move made by hand.
TASK_BOUNDARY = re.compile(
    r"new-task\.sh|git\s+checkout\s+(-q\s+)?-b|git\s+checkout\s+(-q\s+)?task/|git\s+switch\s+(-c\s+)?",
)


def main() -> int:
    try:
        data = json.load(sys.stdin)
    except Exception:
        return 0

    # Exempt subagent sessions. The hook exists to push the main
    # (Opus) session to delegate; once inside a subagent, the
    # subagent IS the delegation and needs to do executor work
    # freely. Claude Code sets agent_id / agent_type on tool calls
    # made inside a subagent invocation (session_id is shared with
    # the parent, so it can't be used as the exemption signal).
    if data.get("agent_id") or data.get("agent_type"):
        return 0

    tool = data.get("tool_name", "")
    tool_input = data.get("tool_input") or {}
    session_id = data.get("session_id", "default")

    # Reset on subagent spawn.
    if tool in ("Task", "Agent"):
        reset(session_id)
        return 0

    is_search = tool in ("Read", "Grep", "Glob")
    cmd = tool_input.get("command", "") if tool == "Bash" else ""
    if tool == "Bash":
        if TASK_BOUNDARY.search(cmd):
            reset(session_id)
            return 0
        if SEARCH_VERBS.match(cmd):
            is_search = True

    if not is_search:
        return 0

    # Files with uncommitted changes are exempt -- neither counted nor blocked.
    # Reading a file you are in the middle of changing is the tail of a known
    # edit, not exploration. For Bash the path has to appear literally in the
    # command, which is how `sed -n 40,60p <file>` on a live file reads as
    # follow-up.
    dirty = dirty_paths(data.get("cwd", ""))
    if dirty:
        target = tool_input.get("file_path") or tool_input.get("path") or ""
        if target and any(target.endswith(d) for d in dirty):
            return 0
        if cmd and any(d in cmd for d in dirty):
            return 0

    n = read_count(session_id) + 1
    write_count(session_id, n)

    if n > THRESHOLD:
        msg = (
            f"{n} executor-style lookups on this one instruction — consider delegating "
            "instead: Explore for research, implementer for scoped mechanical edits "
            "(NOT general-purpose: implementer has no Agent tool, so it cannot nest). "
            "Approve to continue inline. The counter resets on an Agent spawn, on the "
            "next user prompt, and on starting a task branch; reading a file with "
            "uncommitted changes never counts."
        )
        print(json.dumps({
            "hookSpecificOutput": {
                "hookEventName": "PreToolUse",
                "permissionDecision": "ask",
                "permissionDecisionReason": msg,
            }
        }))
    return 0


if __name__ == "__main__":
    sys.exit(main())
