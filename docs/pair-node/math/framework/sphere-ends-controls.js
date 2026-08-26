const SPHERE_END_ANGLES = ['theta', 'phi'];
const SPHERE_END_LABELS = { theta: 'θ top/bottom', phi: 'φ top/bottom' };
const SPHERE_TILT_LABELS = { theta: 'θ tilt pair', phi: 'φ tilt pair' };
const SPHERE_TILT_STEPS = 48;

function endTiltSlider(svgEl, which, onChange) {
  const wrap = document.createElement('label');
  wrap.className = 'endtilt';

  const text = document.createElement('span');
  text.textContent = `tilt ${which === 'theta' ? 'θ' : 'φ'}`;

  const range = document.createElement('input');
  range.type = 'range';
  range.min = 0;
  range.max = SPHERE_TILT_STEPS;
  range.value = 0;
  range.addEventListener('input', () => {
    onChange(Number(range.value) / SPHERE_TILT_STEPS * 2 * Math.PI);
  });

  wrap.appendChild(text);
  wrap.appendChild(range);
  return { wrap, range };
}

function sphereEndControls(svgEl) {
  const row = document.createElement('div');
  row.className = 'ringtoggles';

  svgEl.view.ends = { theta: true, phi: true };
  svgEl.view.tilts = { theta: true, phi: true };
  svgEl.view.endTilt = { theta: 0, phi: 0 };

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

    const s = endTiltSlider(svgEl, which, (turn) => {
      svgEl.view.endTilt[which] = turn;
      svgEl.redraw();
    });
    row.appendChild(s.wrap);
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
