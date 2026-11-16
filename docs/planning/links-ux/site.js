(function () {
  const PAGES = [
    ["index.html", "Overview"],
    ["editing.html", "Editing a link"],
    ["around.html", "Around the board"],
  ];
  const here = location.pathname.split("/").pop() || "index.html";
  const nav = document.getElementById("nav");
  if (nav) {
    nav.innerHTML = "<b>Links tab · UX</b>" + PAGES.map(([f, t], i) =>
      `<a href="${f}" class="${f === here ? "here" : ""}"><small>${String(i).padStart(2, "0")}</small>${t}</a>`).join("");
  }
  const i = PAGES.findIndex(([f]) => f === here);
  const main = document.querySelector("main");
  if (main && i >= 0) {
    const prev = PAGES[i - 1], next = PAGES[i + 1];
    const p = document.createElement("div");
    p.className = "pager";
    p.innerHTML = (prev ? `<a href="${prev[0]}">← ${prev[1]}</a>` : "<span></span>") + (next ? `<a href="${next[0]}">${next[1]} →</a>` : "<span></span>");
    main.appendChild(p);
  }

  const NS = "http://www.w3.org/2000/svg";
  const COL = { phi: "#FF9F1C", theta: "#3DDC84", start: "#9aa0a8", ink: "#222", accent: "#1565c0", guide: "#dde1e6" };
  function el(tag, attrs, parent) {
    const e = document.createElementNS(NS, tag);
    for (const k in attrs) e.setAttribute(k, attrs[k]);
    if (parent) parent.appendChild(e);
    return e;
  }
  function pt(c, r, tick, ccw) {
    const a = (ccw ? -tick : tick) * Math.PI / 6;
    return [c + r * Math.sin(a), c - r * Math.cos(a)];
  }
  function arc(c, r, a, b, ccw) {
    const [x0, y0] = pt(c, r, a, ccw), [x1, y1] = pt(c, r, b, ccw);
    const span = Math.abs(b - a);
    if (span < 1e-6) return "";
    const large = span > 6 ? 1 : 0;
    const sweep = (b > a) !== !!ccw ? 1 : 0;
    return `M${x0},${y0} A${r},${r} 0 ${large} ${sweep} ${x1},${y1}`;
  }
  window.drawDial = function (host) {
    host.innerHTML = "";
    const size = +(host.dataset.size || 20), c = size / 2, r = c - 1.5;
    const start = +(host.dataset.start || 0), turn = +(host.dataset.turn || 0);
    const ghost = host.dataset.ghost !== undefined ? +host.dataset.ghost : null;
    const axis = host.dataset.axis || "phi", ccw = host.dataset.dir === "ccw";
    const col = COL[axis] || COL.phi;
    const svg = el("svg", { width: size, height: size, viewBox: `0 0 ${size} ${size}` }, host);
    el("circle", { cx: c, cy: c, r, fill: "#fff", stroke: COL.guide, "stroke-width": 1 }, svg);
    if (size >= 40) {
      for (let t = 0; t < 12; t++) {
        const [x0, y0] = pt(c, r, t), [x1, y1] = pt(c, r - (t % 3 ? 3 : 6), t);
        el("line", { x1: x0, y1: y0, x2: x1, y2: y1, stroke: "#c8ccd2", "stroke-width": 1 }, svg);
      }
    }
    if (host.dataset.labels !== undefined && size >= 70) {
      for (const t of [0, 3, 6, 9]) {
        const [x, y] = pt(c, r - 13, t);
        const tx = el("text", { x, y: y + 4, "text-anchor": "middle", "font-size": 10, fill: "#555", "font-family": "system-ui" }, svg);
        tx.textContent = t;
      }
    }
    const sw = Math.max(2, size / 14);
    if (ghost !== null && ghost !== turn) {
      el("path", { d: arc(c, r - sw, start, start + ghost, ccw), fill: "none", stroke: col, "stroke-width": sw, "stroke-dasharray": "3 3", opacity: .7 }, svg);
    }
    if (turn) el("path", { d: arc(c, r - sw, start, start + turn, ccw), fill: "none", stroke: col, "stroke-width": sw, "stroke-linecap": "round" }, svg);
    const [sx, sy] = pt(c, r - sw * 1.5, start, ccw);
    el("line", { x1: c, y1: c, x2: sx, y2: sy, stroke: COL.start, "stroke-width": Math.max(1.2, size / 30) }, svg);
    if (turn) {
      const [ex, ey] = pt(c, r - sw * 1.5, start + turn, ccw);
      el("line", { x1: c, y1: c, x2: ex, y2: ey, stroke: COL.ink, "stroke-width": Math.max(1.5, size / 24), "stroke-linecap": "round" }, svg);
      if (size >= 40) el("circle", { cx: ex, cy: ey, r: size / 16, fill: COL.accent }, svg);
    }
    if (ghost !== null && ghost !== turn && size >= 40) {
      const [gx, gy] = pt(c, r - sw * 1.5, start + ghost, ccw);
      el("circle", { cx: gx, cy: gy, r: size / 16, fill: "#fff", stroke: COL.accent, "stroke-width": 1.5 }, svg);
    }
    el("circle", { cx: c, cy: c, r: Math.max(1.2, size / 28), fill: COL.ink }, svg);
  };
  document.querySelectorAll(".dial").forEach(d => { d.style.display = "inline-block"; d.style.lineHeight = 0; d.style.verticalAlign = "middle"; drawDial(d); });
})();
