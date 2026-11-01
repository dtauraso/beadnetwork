const CHECKER_FILES = [
  { file: "rules/numbers.js", what: "the constants and the arithmetic every rule uses",
    parts: () => [`const TAU = 2 * Math.PI;\nconst ROUNDING = ${ROUNDING};\nconst SAME = ${SAME};\nconst AXES = ${JSON.stringify(AXES)};`,
      turn, wrap, add, distance, size, samePlace, sameTicks] },
  { file: "rules/one-state.js", what: "rules 1, 3, 4 and 5, checked on one saved state",
    parts: () => [wherePoints, zFlips, linkVector, nodePosition, linkCloses, checkSequence] },
  { file: "rules/press.js", what: "rules 2 and 6: a step, and what a press predicts",
    parts: () => [stepDelta, pressLink, predict, compare] },
  { file: "rules/order.js", what: "rule 7: Δ does not depend on the order of presses",
    parts: () => [startRewritten, sameButtons, orderCheck] },
];

function codeView() {
  $("source").replaceChildren(...CHECKER_FILES.map(f => el("div", {},
    el("h3", {}, f.file), el("p", { class: "muted" }, f.what),
    el("pre", { class: "code" }, f.parts().map(String).join("\n\n")))));
  for (const pre of document.querySelectorAll("pre[data-fn]")) {
    pre.textContent = pre.dataset.fn.split(" ").map(n => String(window[n])).join("\n\n");
  }
}
