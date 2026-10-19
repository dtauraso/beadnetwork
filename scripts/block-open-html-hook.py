#!/usr/bin/env python3
"""PreToolUse(Bash) hook: block `open` of an .html file.

Rationale: HTML artifacts should render inside VS Code (Live Preview), not the
system browser. The terminal cannot trigger Live Preview's pane directly, but it
CAN reliably keep HTML out of the browser by forcing `code <file>` instead of
`open <file>`. Memory/preference is discipline; this hook is enforcement.

Denies when the command invokes the `open` binary with an argument ending in
.html (with optional #fragment/?query). Arguments are split the way the shell
splits them, so quoting part of a path does not hide the .html. One exception:
`open -a Safari <file>` names the browser deliberately and passes. Everything
else passes through.
"""
import json
import re
import shlex
import sys

try:
    data = json.load(sys.stdin)
except Exception:
    sys.exit(0)

cmd = (data.get("tool_input") or {}).get("command", "")

HTML = re.compile(r"\.html(?:[?#].*)?$")
SEPARATORS = {";", "&", "&&", "|", "||"}


def segments(command):
    lexer = shlex.shlex(command, posix=True, punctuation_chars=";&|")
    lexer.whitespace_split = True
    lexer.commenters = ""
    words = []
    for token in lexer:
        if token in SEPARATORS:
            yield words
            words = []
        else:
            words.append(token)
    yield words


def blocked(words):
    if not words or words[0] != "open":
        return False
    args = words[1:]
    for flag, value in zip(args, args[1:]):
        if flag == "-a" and value.lower() in ("safari", "safari.app"):
            return False
    return any(HTML.search(arg) for arg in args)


try:
    deny = any(blocked(words) for words in segments(cmd.replace("\n", " ; ")))
except ValueError:
    deny = re.search(r"(?:^|[;&|])\s*open\b.*\.html", cmd) is not None

if deny:
    msg = (
        "Blocked: don't `open` HTML in the browser. Use `code \"<file>\"` so it "
        "opens in VS Code, then render it with the Live Preview icon (or cmd+k v). "
        "Re-run with `code` instead of `open`, or `open -a Safari` to choose "
        "Safari deliberately."
    )
    print(json.dumps({
        "hookSpecificOutput": {
            "hookEventName": "PreToolUse",
            "permissionDecision": "deny",
            "permissionDecisionReason": msg,
        }
    }))

sys.exit(0)
