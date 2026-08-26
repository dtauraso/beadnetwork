const SPHERE_CARD_SPEC = {
  points: 12,
  size: 300,
  phi: { axis: 3, arrival: 5 },
  theta: { axis: 0, arrival: 1 },
};

const SPHERE_CARD_FORMULAS = String.raw`\[
\begin{array}{@{}l@{\;}c@{\;}l@{}}
\text{center} &=& (c_{\varphi},\, c_{\theta},\, c_{r}) \\[3pt]
\tau_{\varphi} &=& \text{the whole turn on } \varphi \\[3pt]
\tau_{\theta} &=& \text{the whole turn on } \theta \\[3pt]
\begin{pmatrix} \text{bottom}_{\varphi} \\ \text{bottom}_{\theta} \end{pmatrix}
  &=& \begin{pmatrix} (\text{top}_{\varphi} + \tau_{\varphi}/2) \bmod \tau_{\varphi} \\
                      (\text{top}_{\theta} + \tau_{\theta}/2) \bmod \tau_{\theta} \end{pmatrix} \\[3pt]
\text{distance}_{\text{top}_{\varphi}} &=& |\, \text{top}_{\varphi} - \text{arrival}_{\varphi} \,| \\[3pt]
\text{distance}_{\text{bottom}_{\varphi}} &=& |\, \text{bottom}_{\varphi} - \text{arrival}_{\varphi} \,| \\[3pt]
\text{offset}_{\varphi} &=& 0 \\
& & \quad\text{if } \text{distance}_{\text{top}_{\varphi}} = 0 \text{ or } \text{distance}_{\text{bottom}_{\varphi}} = 0 \\
& & -1 \\
& & \quad\text{if } \text{distance}_{\text{top}_{\varphi}} < \tau_{\varphi}/4 \\
& & +1 \\
& & \quad\text{if } \text{distance}_{\text{bottom}_{\varphi}} < \tau_{\varphi}/4 \\
& & 0 \\
& & \quad\text{otherwise} \\[3pt]
\text{top}_{\text{next}_{\varphi}} &=& (\text{top}_{\varphi} + \text{offset}_{\varphi}) \bmod \tau_{\varphi} \\[3pt]
\text{bottom}_{\text{next}_{\varphi}} &=& (\text{bottom}_{\varphi} + \text{offset}_{\varphi}) \bmod \tau_{\varphi} \\[3pt]
\text{sent}_{\varphi} &=& (\text{top}_{\varphi} + \tau_{\varphi}/4) \bmod \tau_{\varphi} \\[3pt]
\text{normal}_{\varphi} &=& (\text{arrival}_{\varphi} \pm \tau_{\varphi}/4) \bmod \tau_{\varphi} \\
& & \quad\text{the sign that puts it within } \tau_{\varphi}/4 \text{ of } \text{top}_{\varphi} \\[3pt]
\text{point}_{\varphi}(i) &=& \text{the point at index } i \text{ on ring } \varphi \\[3pt]
\text{drawn}_{\varphi}(i) &=& \text{center} + \text{point}_{\varphi}(i)
\end{array}
\]`;

const SPHERE_CARD_INTRO = [
  ['', 'Both angles are in play, \\(\\varphi\\) and \\(\\theta\\). The lines below are written on \\(\\varphi\\); \\(\\theta\\) is the same lines with \\(\\theta\\) in every subscript.'],
];

const SPHERE_CARD_NOTES = [
  ['note', 'Drag either sphere to walk it around the other’s surface — the one you grab moves, the one you don’t stays put, and a grab where they overlap takes the sphere whose centre is nearer. Drag off both to turn the pair. Scroll or shift-drag to pan, pinch (or ctrl-scroll) to zoom, 1 or 2 to turn about that sphere, double-click to put it all back. What faces away is dimmed, not hidden.'],
  ['', 'Every arrow leaves \\(\\text{center}\\), the one thing both rings share. The rule never mentions it: the arithmetic is on indices, and \\(\\text{center}\\) only says where they get drawn.'],
  ['', 'The same rule, once per angle. Each angle carries its own whole turn — \\(\\tau_{\\varphi}\\) and \\(\\tau_{\\theta}\\) — measures its own two distances, and produces its own offset — nothing crosses between \\(\\varphi\\) and \\(\\theta\\), so neither angle can hold the other back. A pair is settled when \\(\\text{offset}_{\\varphi}\\) and \\(\\text{offset}_{\\theta}\\) are both \\(0\\), which is the one-angle halt read on each angle in turn.'],
  ['', '\\(\\text{normal}_{\\varphi}\\) has two candidates a half turn apart; the one named here is the one on \\(\\text{top}_{\\varphi}\\)’s side of the ring.'],
  ['', 'The code in <code>tiltring/rules.go</code> runs this on ONE angle today. This card is the shape it takes when the same arithmetic is carried on the polar lattice’s two angles.'],
];

const SPHERE_RING_LABELS = { theta: 'θ ring', phi: 'φ ring', cross: 'φ×θ ring' };

function ringToggle(svgEl, which, label, onChange) {
  const wrap = document.createElement('label');
  wrap.className = 'ringtoggle';
  const box = document.createElement('input');
  box.type = 'checkbox';
  box.checked = true;
  const text = document.createElement('span');
  text.textContent = label;
  wrap.appendChild(box);
  wrap.appendChild(text);
  box.addEventListener('change', () => onChange(box.checked));
  return { wrap, box };
}

function sphereCardToggles(svgEl) {
  const row = document.createElement('div');
  row.className = 'ringtoggles';
  const boxes = {};

  for (const which of SPHERE_RINGS) {
    const t = ringToggle(svgEl, which, SPHERE_RING_LABELS[which], (on) => {
      svgEl.view.rings[which] = on;
      all.box.checked = SPHERE_RINGS.some((w) => svgEl.view.rings[w]);
      svgEl.redraw();
    });
    boxes[which] = t.box;
    row.appendChild(t.wrap);
  }

  const all = ringToggle(svgEl, 'all', 'all rings', (on) => {
    for (const which of SPHERE_RINGS) {
      svgEl.view.rings[which] = on;
      boxes[which].checked = on;
    }
    svgEl.redraw();
  });
  row.appendChild(all.wrap);
  return row;
}

function sphereCardFigure(spec) {
  const fig = document.createElement('div');
  fig.className = 'keysphere';
  const svgEl = sphere(spec);
  fig.appendChild(svgEl);
  fig.appendChild(sphereCardToggles(svgEl));
  fig.appendChild(sphereEndControls(svgEl));
  return fig;
}

function sphereCardMath(tex) {
  const box = document.createElement('div');
  box.textContent = tex;
  return box;
}

function sphereCardNotes(notes) {
  const box = document.createElement('div');
  box.className = 'keynote';
  for (const [cls, html] of notes) {
    const p = document.createElement('p');
    if (cls) p.className = cls;
    p.innerHTML = html;
    box.appendChild(p);
  }
  return box;
}

for (const host of document.querySelectorAll('[data-sphere-card]')) {
  host.appendChild(sphereCardFigure(SPHERE_CARD_SPEC));
  host.appendChild(sphereCardNotes(SPHERE_CARD_INTRO));
  host.appendChild(sphereCardMath(SPHERE_CARD_FORMULAS));
  host.appendChild(sphereCardNotes(SPHERE_CARD_NOTES));
}
