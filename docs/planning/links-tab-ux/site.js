(function () {
  const PAGES = [
    {
    "file": "index.html",
    "title": "Overview",
    "blurb": ""
    },
    {
    "file": "today.html",
    "title": "Today",
    "blurb": "The current Links tab, row by row, with what gets in the way. Every point is read from pill_links.go and its neighbours."
    },
    {
    "file": "model.html",
    "title": "Model",
    "blurb": "What the user is really editing. Seen as data, a sequence is a links × steps matrix, and that shape drives the whole layout."
    },
    {
    "file": "layout.html",
    "title": "Layout",
    "blurb": "The whole panel: a scope bar, a steps × links score, an inspector for the selected cell, and a status strip. Four regions, one job each."
    },
    {
    "file": "score.html",
    "title": "Score",
    "blurb": "The steps × links grid. Each cell shows its turn as a mini clock, the pressed delta, and whether the link is picked."
    },
    {
    "file": "inspector.html",
    "title": "Inspector",
    "blurb": "Edits the selected cell. The numbers are a clock: a draggable hand that snaps to the increment, with steppers and every value labeled."
    },
    {
    "file": "rules.html",
    "title": "Rules",
    "blurb": "Every rule the model enforces, moved from a refusal after the fact to something the controls make impossible or explain up front."
    },
    {
    "file": "flows.html",
    "title": "Flows",
    "blurb": "Storyboards for the common tasks. Turning a link a quarter turn further takes 4 actions instead of about 9."
    },
    {
    "file": "keyboard.html",
    "title": "Keyboard",
    "blurb": "A tracker-style key map: arrows walk the grid, Space picks, + and − press, brackets change step."
    },
    {
    "file": "states.html",
    "title": "States",
    "blurb": "Empty, single node, many links, nodes falling behind, reloading after a frame edit, and a step in progress."
    },
    {
    "file": "build.html",
    "title": "Build",
    "blurb": "How it maps onto the existing Go and draw code, the new Sheet primitives it needs, and an order that ships one piece at a time."
    }
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
  tabs.querySelector('[aria-current]')?.scrollIntoView({ block: "nearest", inline: "nearest" });
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
    const labels = svg.dataset.labels !== undefined;
    const mini = size < 34;
    const c = size / 2, r = c - (labels ? 13 : mini ? 2 : 5);
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
    el(svg, "circle", { class: "track", cx: c, cy: c, r: r, "stroke-width": mini ? 1.5 : 1 });
    if (!mini) {
      for (let u = 0; u < 12; u++) {
        const major = u % 3 === 0;
        const [x0, y0] = at(u, r - (major ? 4 : 2.5)), [x1, y1] = at(u, r + (major ? 4 : 2.5));
        el(svg, "line", { class: "tick" + (major ? " major" : ""), x1: x0, y1: y0, x2: x1, y2: y1, "stroke-width": major ? 1.3 : 1 });
      }
    }
    if (labels) {
      [0, 3, 6, 9].forEach((u) => {
        const [x, y] = at(u, r + 9);
        const t = el(svg, "text", { x: x, y: y });
        t.textContent = u;
      });
    }
    const w = mini ? 2.5 : size > 100 ? 5 : 3.5;
    if (start) el(svg, "path", { class: "start", d: arc(0, start, r - (mini ? 0 : w + 1)), "stroke-width": mini ? 1.5 : 2 });
    if (delta) el(svg, "path", { class: "delta" + (delta < 0 ? " neg" : ""), d: arc(start, delta, r), "stroke-width": w });
    if (ghost !== null) el(svg, "path", { class: "ghost", d: arc(start + delta, ghost, r), "stroke-width": w });
    if (inc && delta && !mini) {
      const n = Math.abs(delta / inc);
      for (let k = 1; k < n; k++) {
        const [x, y] = at(start + Math.sign(delta) * inc * k);
        el(svg, "circle", { class: "inc", cx: x, cy: y, r: 1.3 });
      }
    }
    const end = start + delta;
    const [ex, ey] = at(end, r * (mini ? 0.95 : 0.78));
    el(svg, "line", { class: "ptr", x1: c, y1: c, x2: ex, y2: ey, "stroke-width": mini ? 1.2 : 1.4 });
    el(svg, "circle", { class: "hub", cx: c, cy: c, r: mini ? 1.2 : 2 });
    if (!mini && start) {
      const [sx, sy] = at(start);
      el(svg, "circle", { class: "startmark", cx: sx, cy: sy, r: 2.6 });
    }
    const [hx, hy] = at(end);
    el(svg, "circle", { class: "end", cx: hx, cy: hy, r: mini ? 2.2 : size > 100 ? 6 : 4 });
    if (ghost !== null) {
      const [gx, gy] = at(end + ghost);
      el(svg, "circle", { class: "ghostend", cx: gx, cy: gy, r: mini ? 2.2 : size > 100 ? 6 : 4, "stroke-width": 1.5 });
    }
  }
  window.__drawDials = () => document.querySelectorAll("svg[data-dial]").forEach((s) => { if (!s.childNodes.length) drawDial(s); });
  window.__drawDials();
})();
