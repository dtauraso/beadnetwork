import esbuild from "esbuild";
import { rename, writeFile } from "node:fs/promises";

const watch = process.argv.includes("--watch");

// esbuild writes an outfile in place: truncate, then stream a megabyte into it.
// The webview loads that file by URL, and a page that loads mid-write gets a
// partial or empty script - a blank editor with no error anywhere, because a
// bundle that never runs cannot report that it never ran. Build in memory and
// put it in place with a rename, which is atomic: a reader sees the whole old
// file or the whole new one. Same reason valuefile writes tmp-then-rename.
const atomicWrite = {
  name: "atomic-write",
  setup(build) {
    build.onEnd(async (result) => {
      const files = result.outputFiles ?? [];
      // Maps first: when the .js appears, everything it points at is already there.
      const ordered = [...files].sort((a, b) => Number(b.path.endsWith(".map")) - Number(a.path.endsWith(".map")));
      for (const file of ordered) {
        const tmp = `${file.path}.tmp`;
        await writeFile(tmp, file.contents);
        await rename(tmp, file.path);
      }
    });
  },
};

const common = {
  bundle: true,
  sourcemap: true,
  target: "es2022",
  logLevel: "info",
  minify: !watch,
  write: false,
  plugins: [atomicWrite],
};

const extension = {
  ...common,
  entryPoints: ["src/extension/extension.ts"],
  outfile: "out/extension.js",
  platform: "node",
  format: "cjs",
  external: ["vscode"],
};

const webview = {
  ...common,
  entryPoints: ["src/webview/main.tsx"],
  outfile: "out/webview.js",
  platform: "browser",
  format: "iife",
  jsx: "automatic",
  loader: { ".css": "css" },

};

if (watch) {
  const ctxA = await esbuild.context(extension);
  const ctxB = await esbuild.context(webview);
  await Promise.all([ctxA.watch(), ctxB.watch()]);
} else {
  await Promise.all([esbuild.build(extension), esbuild.build(webview)]);
}
