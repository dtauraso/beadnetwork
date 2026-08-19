import esbuild from "esbuild";
import { fileURLToPath } from "url";
import path from "path";
import os from "os";
import fs from "fs";
import { execFileSync } from "child_process";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(__dirname, "../../..");
const outPath = process.argv[2] ?? path.join(repoRoot, "src/Node/Wiring/testdata/input_fixture.json");

const bundlePath = path.join(os.tmpdir(), `gen-input-fixture-${process.pid}.cjs`);

await esbuild.build({
  entryPoints: [path.join(__dirname, "gen-input-fixture-src.ts")],
  bundle: true,
  platform: "node",
  format: "cjs",
  outfile: bundlePath,
  logLevel: "warning",
});

try {
  execFileSync(process.execPath, [bundlePath, outPath], { stdio: "inherit" });
} finally {
  fs.rmSync(bundlePath, { force: true });
  fs.rmSync(bundlePath + ".map", { force: true });
}
