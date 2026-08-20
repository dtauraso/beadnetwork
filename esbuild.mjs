import esbuild from "esbuild";

const watch = process.argv.includes("--watch");

const common = {
  bundle: true,
  sourcemap: true,
  target: "es2022",
  logLevel: "info",
  minify: !watch,
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
