"""Read back what Go emitted into a capture-stream-frames.sh directory.

The point of this over recomputing the geometry from the JSON tree: it can DISAGREE.
A check that runs the same arithmetic on both sides is a tautology and will report
success while the screen shows otherwise.

Each write is [len u32][frame]; a frame is [tick u32][labelLen u32][row][label][events].
Strides come from BUF_LAYOUT_FINGERPRINT in
tools/topology-vscode/src/Buffer/buffer-layout.ts -- if that moves and
these do not, the decode reports nonsense, so it is checked below.
"""

import glob
import math
import os
import re
import struct
import sys

HDR = 8


def strides(repo_root):
    path = os.path.join(
        repo_root, "tools/topology-vscode/src/Buffer/buffer-layout.ts"
    )
    with open(path, encoding="utf8") as fh:
        text = fh.read(8192)
    out = {}
    for block in ("Node", "Edge"):
        m = re.search(r"block:%s\[[^\]]*\]:stride:(\d+)" % block, text)
        if not m:
            raise SystemExit(
                "decode-stream-frames: no %s stride in the fingerprint — refusing to "
                "guess a layout" % block
            )
        out[block] = int(m.group(1))
    return out


def frames(path):
    with open(path, "rb") as fh:
        buf = fh.read()
    out, off = [], 0
    while off + 4 <= len(buf):
        (n,) = struct.unpack_from("<I", buf, off)
        off += 4
        if n == 0 or off + n > len(buf):
            break
        out.append(buf[off : off + n])
        off += n
    return out


def main():
    if len(sys.argv) < 2:
        raise SystemExit("usage: decode-stream-frames.py <capture-dir>")
    capture = sys.argv[1]
    repo_root = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
    stride = strides(repo_root)

    nodes = {}
    for path in sorted(glob.glob(os.path.join(capture, "node*.bin"))):
        got = frames(path)
        if not got:
            continue
        frame = got[-1]
        tick, llen = struct.unpack_from("<II", frame, 0)
        row = frame[HDR : HDR + stride["Node"]]
        label = frame[HDR + stride["Node"] : HDR + stride["Node"] + llen].decode(
            "utf8", "replace"
        )
        nid = struct.unpack_from("<i", row, 0)[0]
        nodes[nid] = (tick, struct.unpack_from("<fff", row, 4), label)

    print("NODE CENTRES, as Go emitted them")
    for nid in sorted(nodes):
        tick, (cx, cy, cz), label = nodes[nid]
        print(f"  node {nid} ({label}) tick {tick}: ({cx:.2f}, {cy:.2f}, {cz:.2f})")

    print("\nEDGE ENDS, as Go emitted them")
    print("  an end is pulled back from its target's centre by that node's torus radius,")
    print("  which is a WHOLE number of 8.96 steps. A fractional gap means the edge is")
    print("  aimed somewhere the node is not.\n")
    worst = 0.0
    for path in sorted(glob.glob(os.path.join(capture, "edge*.bin"))):
        got = frames(path)
        if not got:
            continue
        frame = got[-1]
        tick, llen = struct.unpack_from("<II", frame, 0)
        row = frame[HDR : HDR + stride["Edge"]]
        label = frame[HDR + stride["Edge"] : HDR + stride["Edge"] + llen].decode(
            "utf8", "replace"
        )
        ex, ey, ez = struct.unpack_from("<fff", row, 12)
        dst = struct.unpack_from("<i", row, 28)[0]
        target = nodes.get(dst + 1)
        if target is None:
            print(f"  {label:>8}: dstRow {dst} — no node frame captured")
            continue
        ttick, (cx, cy, cz), _ = target
        gap = math.dist((ex, ey, ez), (cx, cy, cz))
        steps = gap / 8.96
        off_grid = abs(steps - round(steps)) > 0.05
        worst = max(worst, abs(steps - round(steps)))
        flag = "OFF-GRID" if off_grid else "ok"
        drift = "" if ttick == tick else f"  [tick {tick} vs node {ttick} — not comparable]"
        print(f"  {label:>8}: gap {gap:7.2f} = {steps:6.2f} steps  {flag}{drift}")

    print(f"\nworst distance from a whole step = {worst:.3f} (0 is exact)")


if __name__ == "__main__":
    main()
