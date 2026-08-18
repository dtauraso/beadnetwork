import esbuild from "esbuild";

const watch = process.argv.includes("--watch");

const common = {
  bundle: true,
  sourcemap: true,
  target: "es2022",
  logLevel: "info",
  minify: !watch,

  // Slider/, Scenes/ and PolarRules/ keep their TS beside their Go at the repo
  // root, so those files sit above this node_modules rather than under it.
  // Node resolution only looks upward, so react and friends need naming here.
  nodePaths: ["node_modules"],
};

const extension = {
  ...common,
  entryPoints: ["src/extension.ts"],
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
