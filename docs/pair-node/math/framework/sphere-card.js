const SPHERE_CARD_SPEC = {
  points: 12,
  size: 300,
  phi: { axis: 3, arrival: 5 },
  theta: { axis: 0, arrival: 1 },
};

const SPHERE_CARD_SHARED = String.raw`\[
\begin{array}{@{}l@{\;}c@{\;}l@{}}
\text{center} &=& \text{where the sphere sits} \\
& & \text{the point every arrow leaves} \\[3pt]
\text{center} &=& (c_{\varphi},\, c_{\theta},\, c_{r}) \\[3pt]
\tau_{\theta} &=& \text{the whole turn on } \theta \\
& & \quad\theta \text{ is an azimuth: a whole turn is a whole turn} \\[3pt]
\tau_{\varphi} &=& \tau_{\theta}/2 \\
& & \quad\varphi \text{ is a colatitude, read through } \sin\varphi,\ \cos\varphi \\
& & \quad(\varphi,\, \theta) \text{ and } (\tau_{\theta} - \varphi,\, \theta + \tau_{\theta}/2) \\
& & \quad\text{are the SAME point, so } \varphi\text{'s own turn is half} \\[3pt]
c_{r} &=& \text{held — nothing below changes it}
\end{array}
\]`;

const SPHERE_CARD_RULE_PHI = String.raw`\[
\begin{array}{@{}l@{\;}c@{\;}l@{}}
\text{top}_{\varphi} &=& 0,\, \tau_{\varphi} - 1 \\
& & \quad\text{the one end the node holds} \\[3pt]
\text{bottom}_{\varphi} &=& (\text{top}_{\varphi} + \tau_{\varphi}/2) \bmod \tau_{\varphi} \\[3pt]
\text{arrival}_{\varphi} &=& 0,\, \tau_{\varphi} - 1 \\
& & \quad\text{the } \varphi \text{ that just came in} \\[3pt]
\text{distance}_{\text{top}_{\varphi}} &=& |\, \text{top}_{\varphi} - \text{arrival}_{\varphi} \,| \\[3pt]
\text{distance}_{\text{bottom}_{\varphi}} &=& |\, \text{bottom}_{\varphi} - \text{arrival}_{\varphi} \,| \\[3pt]
\text{offset}_{\varphi} &=& \begin{array}{@{}l@{\;}l@{}}
      0 & \begin{array}{@{}l@{}} \text{if } \text{distance}_{\text{top}_{\varphi}} = 0 \\ \text{and } \text{distance}_{\text{bottom}_{\varphi}} = 0 \end{array} \\[8pt]
      -1 & \text{if } \text{distance}_{\text{top}_{\varphi}} < \tau_{\varphi}/4 \\
      1 & \text{if } \text{distance}_{\text{bottom}_{\varphi}} < \tau_{\varphi}/4 \\
      0 & \text{otherwise}
    \end{array} \\[3pt]
\text{sent}_{\varphi} &=& (c_{\varphi} + \text{offset}_{\varphi}) \bmod \tau_{\varphi}
\end{array}
\]`;

const SPHERE_CARD_RULE_THETA = String.raw`\[
\begin{array}{@{}l@{\;}c@{\;}l@{}}
\text{top}_{\theta} &=& 0,\, \tau_{\theta} - 1 \\
& & \quad\text{the one end the node holds} \\[3pt]
\text{bottom}_{\theta} &=& (\text{top}_{\theta} + \tau_{\theta}/2) \bmod \tau_{\theta} \\[3pt]
\text{arrival}_{\theta} &=& 0,\, \tau_{\theta} - 1 \\
& & \quad\text{the } \theta \text{ that just came in} \\[3pt]
\text{distance}_{\text{top}_{\theta}} &=& |\, \text{top}_{\theta} - \text{arrival}_{\theta} \,| \\[3pt]
\text{distance}_{\text{bottom}_{\theta}} &=& |\, \text{bottom}_{\theta} - \text{arrival}_{\theta} \,| \\[3pt]
\text{offset}_{\theta} &=& \begin{array}{@{}l@{\;}l@{}}
      0 & \begin{array}{@{}l@{}} \text{if } \text{distance}_{\text{top}_{\theta}} = 0 \\ \text{and } \text{distance}_{\text{bottom}_{\theta}} = 0 \end{array} \\[8pt]
      -1 & \text{if } \text{distance}_{\text{top}_{\theta}} < \tau_{\theta}/4 \\
      1 & \text{if } \text{distance}_{\text{bottom}_{\theta}} < \tau_{\theta}/4 \\
      0 & \text{otherwise}
    \end{array} \\[3pt]
\text{sent}_{\theta} &=& (c_{\theta} + \text{offset}_{\theta}) \bmod \tau_{\theta}
\end{array}
\]`;

const SPHERE_CARD_INTRO = [
  ['', 'Both angles are in play, \\(\\varphi\\) and \\(\\theta\\), and each one runs the rule SEPARATELY — its own ends, its own distances, its own offset, its own whole turn. Nothing crosses between them.'],
  ['', 'They are not the same ring. Written as one column vector with one \\(\\tau\\), the rule only holds where \\(\\varphi\\) and \\(\\theta\\) happen to agree: \\(\\theta\\) is an azimuth and its whole turn is a whole turn, while \\(\\varphi\\) is a colatitude the position is read through as \\(\\sin\\varphi, \\cos\\varphi\\), so \\((\\varphi, \\theta)\\) and \\((\\tau_{\\theta} - \\varphi, \\theta + \\tau_{\\theta}/2)\\) name ONE point and \\(\\varphi\\)’s own whole turn is half of \\(\\theta\\)’s.'],
];

const SPHERE_CARD_PHI_HEAD = [
  ['', 'The \\(\\varphi\\) rule, on \\(\\varphi\\)’s own ring:'],
];

const SPHERE_CARD_THETA_HEAD = [
  ['', 'The \\(\\theta\\) rule, on \\(\\theta\\)’s own ring — the same shape, a different whole turn:'],
];

const SPHERE_CARD_NOTES = [
  ['note', 'Drag either sphere to walk it around the other’s surface — the one you grab moves, the one you don’t stays put, and a grab where they overlap takes the sphere whose center is nearer. Drag off both to turn the pair. Scroll or shift-drag to pan, pinch (or ctrl-scroll) to zoom, 1 or 2 to turn about that sphere, double-click to put it all back. What faces away is dimmed, not hidden.'],
  ['', 'Every arrow leaves \\(\\text{center}\\), the one thing both rings share. Each rule steps its own coordinate of it and sends that on; \\(c_{r}\\) is carried through untouched, so a step turns the pair without changing how far apart they are.'],
  ['', 'The code in <code>tiltring/rules.go</code> runs this on ONE angle. <code>Categories/NodeKinds/NodePhiTheta</code> runs both, as two separate rings — <code>column_vectors.go</code> is this card, with \\(\\varphi\\)’s whole turn set to half of \\(\\theta\\)’s.'],
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
  host.appendChild(sphereCardMath(SPHERE_CARD_SHARED));
  host.appendChild(sphereCardNotes(SPHERE_CARD_PHI_HEAD));
  host.appendChild(sphereCardMath(SPHERE_CARD_RULE_PHI));
  host.appendChild(sphereCardNotes(SPHERE_CARD_THETA_HEAD));
  host.appendChild(sphereCardMath(SPHERE_CARD_RULE_THETA));
  host.appendChild(sphereCardNotes(SPHERE_CARD_NOTES));
}
