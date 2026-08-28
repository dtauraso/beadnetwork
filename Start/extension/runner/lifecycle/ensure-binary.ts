import * as fs from "fs";
import { buildBinary, maxGoMtime } from "../../goBuild";

const BUILD_BINARY_MAX_ATTEMPTS = 50;

export async function ensureBinaryBuilt(
  repoRoot: string,
  binPath: string,
): Promise<{ ok: true } | { ok: false; error: string }> {
  let binMtime = -1;
  try {
    binMtime = fs.statSync(binPath).mtimeMs;
  } catch { /* eslint-disable-line no-empty */ }
  const needsRebuild = binMtime < 0 || maxGoMtime(repoRoot) > binMtime;
  if (!needsRebuild) return { ok: true };

  for (let attempt = 0; attempt < BUILD_BINARY_MAX_ATTEMPTS; attempt++) {
    const res = await buildBinary(repoRoot, binPath);
    if (!res.ok) return res;
    if (!res.busy) {

      if (fs.existsSync(binPath)) return { ok: true };
      return { ok: false, error: `go build reported success but ${binPath} is missing` };
    }

    if (fs.existsSync(binPath)) return { ok: true };

    await new Promise((done) => setTimeout(done, 50));
  }
  return {
    ok: false,
    error: `binary ${binPath} not built after ${BUILD_BINARY_MAX_ATTEMPTS} attempts`,
  };
}
