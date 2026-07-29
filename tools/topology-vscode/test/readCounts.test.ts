// readCounts.test.ts — replaces countNodes.test.ts/countEdges (untested) now that step 6
// of .claude/rules/persistence-ownership.md deleted the tree-walking countNodes/
// countEdges in favor of a single stored `<topologyPath>/counts.json` read (see
// runCommand.ts's readCounts doc comment). Unlike the old functions, which returned 0 on
// any read/parse failure, readCounts THROWS — there is no correct fallback count to guess.

import { describe, it, expect } from "vitest";
import * as fs from "fs";
import * as os from "os";
import * as path from "path";
import { readCounts } from "../src/runCommand";

function mkTmpDir(): string {
  return fs.mkdtempSync(path.join(os.tmpdir(), "wirefold-readcounts-"));
}

describe("readCounts", () => {
  it("reads {nodes, edges} from <topologyPath>/counts.json", () => {
    const root = mkTmpDir();
    fs.writeFileSync(path.join(root, "counts.json"), JSON.stringify({ nodes: 9, edges: 10 }));
    expect(readCounts(root)).toEqual({ nodes: 9, edges: 10 });
  });

  it("throws when counts.json is missing", () => {
    const root = mkTmpDir();
    expect(() => readCounts(root)).toThrow(/cannot read/);
  });

  it("throws on a nonexistent topology path", () => {
    expect(() => readCounts("/definitely/does/not/exist/topology")).toThrow();
  });

  it("throws on unparseable JSON", () => {
    const root = mkTmpDir();
    fs.writeFileSync(path.join(root, "counts.json"), "{not json");
    expect(() => readCounts(root)).toThrow(/cannot parse/);
  });

  it("throws when nodes/edges are missing or non-numeric", () => {
    const root = mkTmpDir();
    fs.writeFileSync(path.join(root, "counts.json"), JSON.stringify({ nodes: 9 }));
    expect(() => readCounts(root)).toThrow(/must be/);
  });

  it("throws on negative or non-integer counts", () => {
    const root = mkTmpDir();
    fs.writeFileSync(path.join(root, "counts.json"), JSON.stringify({ nodes: -1, edges: 10 }));
    expect(() => readCounts(root)).toThrow(/must be/);

    const root2 = mkTmpDir();
    fs.writeFileSync(path.join(root2, "counts.json"), JSON.stringify({ nodes: 1.5, edges: 10 }));
    expect(() => readCounts(root2)).toThrow(/must be/);
  });

  it("reads the committed topology's counts.json (repo self-check)", () => {
    const repoRoot = path.join(__dirname, "..", "..", "..");
    const counts = readCounts(path.join(repoRoot, "topology"));
    expect(counts.nodes).toBeGreaterThan(0);
    expect(counts.edges).toBeGreaterThan(0);
  });
});
