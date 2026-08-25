const SPHERE_CARD_SPEC = {
  points: 12,
  size: 300,
  seat: { which: 'phi', index: 5 },
  phi: { axis: 3, arrival: 5 },
  theta: { axis: 0, arrival: 1 },
};

const SPHERE_CARD_FORMULAS = String.raw`\[
\begin{array}{@{}l@{\;}c@{\;}l@{}}
\text{center} &=& \text{the middle of the sphere} \\[3pt]
a &\in& \{\varphi, \theta\} \\[3pt]
\tau_a &=& \text{the whole turn on } a \\[3pt]
\text{bottom}_a &=& (\text{top}_a + \tau_a/2) \bmod \tau_a \\[3pt]
\text{distance}_{\text{top},a} &=& |\, \text{top}_a - \text{arrival}_a \,| \\[3pt]
\text{distance}_{\text{bottom},a} &=& |\, \text{bottom}_a - \text{arrival}_a \,| \\[3pt]
\text{offset}_a &=& 0 \\
& & \quad\text{if } \text{distance}_{\text{top},a} = 0 \text{ or } \text{distance}_{\text{bottom},a} = 0 \\
& & -1 \\
& & \quad\text{if } \text{distance}_{\text{top},a} < \tau_a/4 \\
& & +1 \\
& & \quad\text{if } \text{distance}_{\text{bottom},a} < \tau_a/4 \\
& & 0 \\
& & \quad\text{otherwise} \\[3pt]
\text{top}_{\text{next},a} &=& (\text{top}_a + \text{offset}_a) \bmod \tau_a \\[3pt]
\text{bottom}_{\text{next},a} &=& (\text{bottom}_a + \text{offset}_a) \bmod \tau_a \\[3pt]
\text{sent}_a &=& (\text{top}_a + \tau_a/4) \bmod \tau_a \\[3pt]
\text{normal}_a &=& (\text{arrival}_a \pm \tau_a/4) \bmod \tau_a \\
& & \quad\text{the sign that puts it within } \tau_a/4 \text{ of } \text{top}_a \\[3pt]
\text{point}_a(i) &=& \text{the point at index } i \text{ on ring } a \\[3pt]
\text{drawn}_a(i) &=& \text{center} + \text{point}_a(i)
\end{array}
\]`;

const SPHERE_CARD_NOTES = [
  ['note', 'Drag the sphere to turn it, scroll to zoom, double-click to put it back. What faces away is dimmed, not hidden.'],
  ['', 'Every arrow leaves \\(\\text{center}\\), the one thing both rings share. The rule never mentions it: the arithmetic is on indices, and \\(\\text{center}\\) only says where they get drawn.'],
  ['', 'The same rule, once per angle. Each angle carries its own whole turn \\(\\tau_a\\), measures its own two distances, and produces its own offset — nothing crosses between \\(\\varphi\\) and \\(\\theta\\), so neither angle can hold the other back. A pair is settled when every \\(\\text{offset}_a\\) is \\(0\\), which is the one-angle halt read on each angle in turn.'],
  ['', '\\(\\text{normal}_a\\) has two candidates a half turn apart; the one named here is the one on \\(\\text{top}_a\\)’s side of the ring.'],
  ['', 'The code in <code>tiltring/rules.go</code> runs this on ONE angle today. This card is the shape it takes when the same arithmetic is carried on the polar lattice’s two angles.'],
];

function sphereCardFigure(spec) {
  const fig = document.createElement('div');
  fig.className = 'keysphere';
  fig.appendChild(sphere(spec));
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
  host.appendChild(sphereCardMath(SPHERE_CARD_FORMULAS));
  host.appendChild(sphereCardNotes(SPHERE_CARD_NOTES));
}
