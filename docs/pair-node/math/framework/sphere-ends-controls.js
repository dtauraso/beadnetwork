const SPHERE_END_ANGLES = ['theta', 'phi'];
const SPHERE_END_LABELS = { theta: 'θ top/bottom', phi: 'φ top/bottom' };
const SPHERE_TILT_LABELS = { theta: 'θ tilt pair', phi: 'φ tilt pair' };

function sphereEndControls(svgEl) {
  const row = document.createElement('div');
  row.className = 'ringtoggles';

  svgEl.view.ends = { theta: true, phi: true };
  svgEl.view.tilts = { theta: true, phi: true };

  const boxes = {};

  for (const which of SPHERE_END_ANGLES) {
    const t = ringToggle(svgEl, which, SPHERE_END_LABELS[which], (on) => {
      svgEl.view.ends[which] = on;
      all.box.checked = SPHERE_END_ANGLES.some((w) => svgEl.view.ends[w]);
      svgEl.redraw();
    });
    boxes[which] = t.box;
    row.appendChild(t.wrap);

    const v = ringToggle(svgEl, which, SPHERE_TILT_LABELS[which], (on) => {
      svgEl.view.tilts[which] = on;
      svgEl.redraw();
    });
    row.appendChild(v.wrap);
  }

  const all = ringToggle(svgEl, 'all', 'all top/bottom', (on) => {
    for (const which of SPHERE_END_ANGLES) {
      svgEl.view.ends[which] = on;
      boxes[which].checked = on;
    }
    svgEl.redraw();
  });
  row.appendChild(all.wrap);

  return row;
}
