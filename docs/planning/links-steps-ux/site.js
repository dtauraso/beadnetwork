/*
branch: task/links-mockup-ui
*/
(function () {
  const PAGES = [
    { file: "index.html", title: "Overview", blurb: "" },
    { file: "steps.html", title: "Steps", blurb: "Each step is a column you open or close on its own. Open as many as you like, side by side; closed ones shrink to a strip." },
    { file: "nodes.html", title: "Nodes", blurb: "Inside an open step, each node is a row that folds on its own, separately per step." },
    { file: "links.html", title: "Links", blurb: "A link row opens its card right where it is. Any number of cards stay open across nodes and steps, each editable." },
    { file: "head.html", title: "HEAD", blurb: "The HEAD row at the top of each open step: a lock per axis and ▲▼ that press every picked link in that step." },
    { file: "diagram.html", title: "Diagram", blurb: "Open cards light their edges in the scene; clicking an edge opens or closes its card. Picked edges are what animates." },
    { file: "rules.html", title: "Rules", blurb: "Picks, fixed r, r at 0, START: what the controls allow, and what they say when a rule applies." },
    { file: "states.html", title: "States", blurb: "Empty, one node, a step running, narrow panel, the Network view and the Frame popover." },
    { file: "build.html", title: "Build", blurb: "What already exists in the model, what the layout code becomes, and the order it is built in." },
  ];
  const here = location.pathname.split("/").pop() || "index.html";
  const tabs = document.getElementById("tabs");
  const i = Math.max(0, PAGES.findIndex((p) => p.file === here));
  PAGES.forEach((p, j) => {
    const a = document.createElement("a");
    a.href = p.file;
    a.innerHTML = '<span class="n">' + String(j + 1).padStart(2, "0") + "</span>" + p.title;
    if (j === i) a.setAttribute("aria-current", "page");
    tabs.appendChild(a);
  });
  const f = document.createElement("nav");
  f.className = "footer-nav";
  const prev = PAGES[i - 1], next = PAGES[i + 1];
  if (prev) f.innerHTML += '<a href="' + prev.file + '"><span class="dir">← Previous</span>' + prev.title + "</a>";
  if (next) f.innerHTML += '<a class="next" href="' + next.file + '"><span class="dir">Next →</span>' + next.title + "</a>";
  document.querySelector("main").appendChild(f);
  const cards = document.getElementById("pagecards");
  if (cards) PAGES.slice(1).forEach((p) => {
    const a = document.createElement("a");
    a.className = "card";
    a.href = p.file;
    a.innerHTML = "<h3>" + p.title + "</h3><p>" + p.blurb + "</p>";
    cards.appendChild(a);
  });

  const sgn = (n) => (n > 0 ? "+" + n : n < 0 ? "−" + -n : "0");
  const stepper = (v, dimDown, dimUp) =>
    '<span class="step"><span class="b' + (dimDown ? " dim" : "") + '">−</span><span class="v">' + v + '</span><span class="b' + (dimUp ? " dim" : "") + '">＋</span></span>';
  function card(div) {
    const nums = (k) => (div.dataset[k] || "0,0,1,0").split(",").map(Number);
    const [ps, pd, pi, pt] = nums("phi"), [ts, td, ti, tt] = nums("theta"), [rs, rd, ri, rt, fixed] = nums("r");
    const dial = (axis, s, d, inc) => '<svg data-dial data-size="36"' + (axis === "theta" ? ' data-axis="theta"' : "") + ' data-start="' + s + '" data-delta="' + d + '" data-inc="' + inc + '"></svg>';
    const end = rs + rt + rd, most = Math.max(rs, end, 1) * 1.3;
    const pct = (v) => Math.max(0, (v / most) * 100);
    const ruler = '<div class="ruler"><div class="t"></div><div class="s" style="width:' + pct(rs) + '%"></div><div class="d" style="left:' + pct(Math.min(rs, end)) + "%;width:" + Math.abs(pct(end) - pct(rs)) + '%"></div><div class="k" style="left:calc(' + pct(end) + '% - 5px)"></div></div>';
    let html = '<span></span><span></span><span class="h">Δ</span><span class="h">nΔ</span><span class="h">total</span>';
    html += dial("phi", ps + pt, pd, pi) + '<span class="ax">φ</span>' + stepper(sgn(pd)) + stepper(pi) + '<span class="tot">' + pt + "</span>";
    html += dial("theta", ts + tt, td, ti) + '<span class="ax">θ</span>' + stepper(sgn(td)) + stepper(ti) + '<span class="tot">' + tt + "</span>";
    html += ruler + '<span class="ax">r</span>' + stepper(sgn(rd), fixed || end <= 0, fixed) + stepper(ri) + '<span class="tot">' + rt + "</span>";
    if (fixed) html += '<div class="lfoot"><span class="chip fixed">🔒 r fixed</span> ' + (div.dataset.why || "") + "</div>";
    else if (end <= 0) html += '<div class="lfoot">r is at 0, the closest it can go</div>';
    div.classList.add("lcard");
    div.innerHTML = html;
  }
  document.querySelectorAll("[data-card]").forEach(card);

  const NS = "http://www.w3.org/2000/svg";
  function el(parent, tag, attrs) {
    const e = document.createElementNS(NS, tag);
    for (const k in attrs) e.setAttribute(k, attrs[k]);
    parent.appendChild(e);
    return e;
  }
  function drawDial(svg) {
    const size = +(svg.dataset.size || 72);
    const start = +(svg.dataset.start || 0);
    const delta = +(svg.dataset.delta || 0);
    const inc = +(svg.dataset.inc || 0);
    const ghost = svg.dataset.ghost !== undefined ? +svg.dataset.ghost : null;
    const mini = size < 30;
    const c = size / 2, r = c - (mini ? 2 : 4);
    svg.classList.add("dial");
    svg.setAttribute("viewBox", "0 0 " + size + " " + size);
    svg.setAttribute("width", size);
    svg.setAttribute("height", size);
    const at = (u, rad) => {
      const t = (u / 12) * 2 * Math.PI;
      return [c + (rad ?? r) * Math.sin(t), c - (rad ?? r) * Math.cos(t)];
    };
    const arc = (from, span, rad) => {
      if (span === 0) return "";
      const s = Math.max(-11.999, Math.min(11.999, span));
      const [x0, y0] = at(from, rad), [x1, y1] = at(from + s, rad);
      return "M" + x0 + " " + y0 + "A" + rad + " " + rad + " 0 " + (Math.abs(s) > 6 ? 1 : 0) + " " + (s > 0 ? 1 : 0) + " " + x1 + " " + y1;
    };
    el(svg, "circle", { class: "track", cx: c, cy: c, r: r, "stroke-width": 1.5 });
    if (!mini) {
      for (let u = 0; u < 12; u += 3) {
        const [x0, y0] = at(u, r - 3), [x1, y1] = at(u, r + 2);
        el(svg, "line", { class: "tick major", x1: x0, y1: y0, x2: x1, y2: y1, "stroke-width": 1 });
      }
    }
    const w = mini ? 2.5 : 3;
    if (start) el(svg, "path", { class: "start", d: arc(0, start, r - (mini ? 0 : w + 1)), "stroke-width": 1.5 });
    if (delta) el(svg, "path", { class: "delta" + (delta < 0 ? " neg" : ""), d: arc(start, delta, r), "stroke-width": w });
    if (ghost !== null) el(svg, "path", { class: "ghost", d: arc(start + delta, ghost, r), "stroke-width": w });
    if (inc && delta && !mini) {
      const n = Math.abs(delta / inc);
      for (let k = 1; k < n; k++) {
        const [x, y] = at(start + Math.sign(delta) * inc * k);
        el(svg, "circle", { class: "inc", cx: x, cy: y, r: 1.2 });
      }
    }
    const end = start + delta;
    const [ex, ey] = at(end, r * 0.8);
    el(svg, "line", { class: "ptr", x1: c, y1: c, x2: ex, y2: ey, "stroke-width": 1.2 });
    el(svg, "circle", { class: "hub", cx: c, cy: c, r: 1.3 });
    const [hx, hy] = at(end);
    el(svg, "circle", { class: "end", cx: hx, cy: hy, r: mini ? 2.2 : 3.2 });
    if (ghost !== null) {
      const [gx, gy] = at(end + ghost);
      el(svg, "circle", { class: "ghostend", cx: gx, cy: gy, r: 3.2, "stroke-width": 1.5 });
    }
  }
  document.querySelectorAll("svg[data-dial]").forEach((s) => { if (!s.childNodes.length) drawDial(s); });
})();
