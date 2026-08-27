const SPHERE_CARD_SPEC = {
  points: 12,
  size: 300,
  phi: { axis: 3, arrival: 5 },
  theta: { axis: 0, arrival: 1 },
};

const SPHERE_CARD_FORMULAS = String.raw`\[
\begin{array}{@{}l@{\;}c@{\;}l@{}}
\text{center} &=& \text{where the sphere sits} \\
& & \text{the point every arrow leaves} \\[3pt]
\text{center} &=& (c_{\varphi},\, c_{\theta},\, c_{r}) \\[3pt]
\tau_{\varphi} &=& \text{the whole turn on } \varphi \\[3pt]
\tau_{\theta} &=& \text{the whole turn on } \theta \\[3pt]
\tau_{r} &=& \text{the whole run on } r \\[3pt]
\begin{bmatrix} \text{top}_{\varphi} \\ \text{top}_{\theta} \end{bmatrix}
  &=& \begin{bmatrix} 0,\, \tau_{\varphi} - 1 \\ 0,\, \tau_{\theta} - 1 \end{bmatrix} \\
& & \quad\text{the one end the node holds} \\
& & \quad\text{an index on each ring} \\[3pt]
\begin{bmatrix} \text{bottom}_{\varphi} \\ \text{bottom}_{\theta} \end{bmatrix}
  &=& \begin{bmatrix} (\text{top}_{\varphi} + \tau_{\varphi}/2) \bmod \tau_{\varphi} \\
                      (\text{top}_{\theta} + \tau_{\theta}/2) \bmod \tau_{\theta} \end{bmatrix} \\[3pt]
\begin{bmatrix} \text{arrival}_{\varphi} \\ \text{arrival}_{\theta} \\ \text{arrival}_{r} \end{bmatrix}
  &=& \begin{bmatrix} 0,\, \tau_{\varphi} - 1 \\ 0,\, \tau_{\theta} - 1 \\ 0,\, \tau_{r} - 1 \end{bmatrix} \\
& & \quad\text{the direction that just came in} \\
& & \quad\text{an index on each ring, like top} \\[3pt]
\begin{bmatrix} \text{distance}_{\text{top}_{\varphi}} \\ \text{distance}_{\text{top}_{\theta}} \end{bmatrix}
  &=& \begin{bmatrix} \min \begin{array}{@{}l@{}}
                             (\ |\, \text{top}_{\varphi} - \text{arrival}_{\varphi} \,|, \\
                             \ \ \tau_{\varphi} - |\, \text{top}_{\varphi} - \text{arrival}_{\varphi} \,|\ )
                           \end{array} \\[10pt]
                      \min \begin{array}{@{}l@{}}
                             (\ |\, \text{top}_{\theta} - \text{arrival}_{\theta} \,|, \\
                             \ \ \tau_{\theta} - |\, \text{top}_{\theta} - \text{arrival}_{\theta} \,|\ )
                           \end{array} \end{bmatrix} \\
& & \quad\text{the way round the ring that is SHORTER} \\
& & \quad\text{the radius vector is decoupled from the axis,} \\
& & \quad\text{so the acute side has to be chosen, not assumed} \\[3pt]
\begin{bmatrix} \text{distance}_{\text{bottom}_{\varphi}} \\ \text{distance}_{\text{bottom}_{\theta}} \end{bmatrix}
  &=& \begin{bmatrix} \min \begin{array}{@{}l@{}}
                             (\ |\, \text{bottom}_{\varphi} - \text{arrival}_{\varphi} \,|, \\
                             \ \ \tau_{\varphi} - |\, \text{bottom}_{\varphi} - \text{arrival}_{\varphi} \,|\ )
                           \end{array} \\[10pt]
                      \min \begin{array}{@{}l@{}}
                             (\ |\, \text{bottom}_{\theta} - \text{arrival}_{\theta} \,|, \\
                             \ \ \tau_{\theta} - |\, \text{bottom}_{\theta} - \text{arrival}_{\theta} \,|\ )
                           \end{array} \end{bmatrix} \\
& & \quad\text{top and bottom are a half turn apart, so these} \\
& & \quad\text{two sum to } \tau/2 \text{ — one is always acute} \\[3pt]
\begin{bmatrix} \text{distance}_{\text{own}_{\varphi}} \\ \text{distance}_{\text{own}_{\theta}} \end{bmatrix}
  &=& \begin{bmatrix} \min \begin{array}{@{}l@{}}
                             (\ |\, c_{\varphi} - \text{top}_{\varphi} \,|, \\
                             \ \ \tau_{\varphi} - |\, c_{\varphi} - \text{top}_{\varphi} \,|\ )
                           \end{array} \\[10pt]
                      \min \begin{array}{@{}l@{}}
                             (\ |\, c_{\theta} - \text{top}_{\theta} \,|, \\
                             \ \ \tau_{\theta} - |\, c_{\theta} - \text{top}_{\theta} \,|\ )
                           \end{array} \end{bmatrix} \\
& & \quad\text{how far this angle's OWN center is from its axis} \\[3pt]
\begin{bmatrix} \text{offset}_{\varphi} \\ \text{offset}_{\theta} \end{bmatrix}
  &=& \begin{bmatrix}
      0 & \text{if } \text{distance}_{\text{own}_{\varphi}} = \tau_{\varphi}/4 \\
      0 &
    \end{bmatrix} \\
& & \quad\text{arrived: an angle stops on ITS OWN quarter turn,} \\
& & \quad\text{so it does not need the other to be there at the} \\
& & \quad\text{same moment, and having stopped it stays} \\
& & \begin{bmatrix}
      0 & \begin{array}{@{}l@{}} \text{if } \text{distance}_{\text{top}_{\varphi}} = 0 \\ \text{and } \text{distance}_{\text{bottom}_{\varphi}} = 0 \end{array} \\[8pt]
      0 &
    \end{bmatrix} \\
& & \begin{bmatrix}
      -1 & \text{if } \text{distance}_{\text{top}_{\varphi}} < \tau_{\varphi}/4 \\
      0 &
    \end{bmatrix} \\
& & \begin{bmatrix}
      1 & \text{if } \text{distance}_{\text{bottom}_{\varphi}} < \tau_{\varphi}/4 \\
      0 &
    \end{bmatrix} \\
& & \quad\text{no OTHERWISE: the distances sum to } \tau/2, \text{ so} \\
& & \quad\text{"neither is acute" is only the quarter-turn} \\
& & \quad\text{boundary — the PARTNER's angle having arrived,} \\
& & \quad\text{which is not this angle's business} \\[3pt]
& & \begin{bmatrix}
      0 & \\
      0 & \text{if } \text{distance}_{\text{own}_{\theta}} = \tau_{\theta}/4
    \end{bmatrix} \\
& & \begin{bmatrix}
      0 & \\[8pt]
      0 & \begin{array}{@{}l@{}} \text{if } \text{distance}_{\text{top}_{\theta}} = 0 \\ \text{and } \text{distance}_{\text{bottom}_{\theta}} = 0 \end{array}
    \end{bmatrix} \\
& & \begin{bmatrix}
      0 & \\
      -1 & \text{if } \text{distance}_{\text{top}_{\theta}} < \tau_{\theta}/4
    \end{bmatrix} \\
& & \begin{bmatrix}
      0 & \\
      1 & \text{if } \text{distance}_{\text{bottom}_{\theta}} < \tau_{\theta}/4
    \end{bmatrix} \\[3pt]
\begin{bmatrix} \text{center}_{\text{next}_{\varphi}} \\ \text{center}_{\text{next}_{\theta}} \\ \text{center}_{\text{next}_{r}} \end{bmatrix}
  &=& \begin{bmatrix} c_{\varphi} + \text{offset}_{\varphi} \\
                      c_{\theta} + \text{offset}_{\theta} \\
                      c_{r} \end{bmatrix} \\
& & \quad\text{no } \bmod: \text{ the center moves by one or not at all,} \\
& & \quad\text{and the LOCK is what stops it — an angle whose} \\
& & \quad\text{offset is } 0 \text{ never moves its center again} \\[3pt]
\begin{bmatrix} \text{sent}_{\varphi} \\ \text{sent}_{\theta} \\ \text{sent}_{r} \end{bmatrix}
  &=& \begin{bmatrix} \text{center}_{\text{next}_{\varphi}} \\
                      \text{center}_{\text{next}_{\theta}} \\
                      \text{center}_{\text{next}_{r}} \end{bmatrix}
\end{array}
\]`;

const SPHERE_CARD_INTRO = [
  ['', 'Both angles are in play, \\(\\varphi\\) and \\(\\theta\\). Each line below is a column, \\(\\varphi\\) over \\(\\theta\\) — the same rule on each row, and nothing crossing between the rows.'],
];

const SPHERE_CARD_NOTES = [
  ['note', 'Drag either sphere to walk it around the other’s surface — the one you grab moves, the one you don’t stays put, and a grab where they overlap takes the sphere whose center is nearer. Drag off both to turn the pair. Scroll or shift-drag to pan, pinch (or ctrl-scroll) to zoom, 1 or 2 to turn about that sphere, double-click to put it all back. What faces away is dimmed, not hidden.'],
  ['', 'Every arrow leaves \\(\\text{center}\\), the one thing both rings share. The rule never mentions it: the arithmetic is on indices, and \\(\\text{center}\\) only says where they get drawn.'],
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
