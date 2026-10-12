function atobCanvas(ctx, W, H, S, view, focus, css) {
  function proj(p) {
    const f = focus(), x = p[0] - f[0], y = p[1] - f[1], z = p[2] - f[2];
    const cy = Math.cos(view.yaw), sy = Math.sin(view.yaw);
    const x1 = x * cy + z * sy, z1 = -x * sy + z * cy;
    const cp = Math.cos(view.pitch), sp = Math.sin(view.pitch);
    const y2 = y * cp - z1 * sp, z2 = y * sp + z1 * cp;
    return [W / 2 + x1 * S, H / 2 - y2 * S, z2];
  }

  function curve(points, about, color, width, alpha = 1) {
    const dc = proj(about)[2];
    ctx.strokeStyle = color; ctx.lineWidth = width;
    for (let i = 0; i + 1 < points.length; i++) {
      const a = proj(points[i]), b = proj(points[i + 1]);
      ctx.globalAlpha = alpha * ((a[2] + b[2]) / 2 < dc ? 0.3 : 1);
      ctx.beginPath(); ctx.moveTo(a[0], a[1]); ctx.lineTo(b[0], b[1]); ctx.stroke();
    }
    ctx.globalAlpha = 1;
  }

  function arrow(p, q, color, width, dashed) {
    const a = proj(p), b = proj(q);
    const ang = Math.atan2(b[1] - a[1], b[0] - a[0]);
    if (Math.hypot(b[0] - a[0], b[1] - a[1]) < 2) return;
    ctx.strokeStyle = color; ctx.fillStyle = color; ctx.lineWidth = width;
    ctx.setLineDash(dashed ? [10, 8] : []);
    ctx.beginPath(); ctx.moveTo(a[0], a[1]); ctx.lineTo(b[0] - 14 * Math.cos(ang), b[1] - 14 * Math.sin(ang)); ctx.stroke();
    ctx.setLineDash([]);
    ctx.beginPath(); ctx.moveTo(b[0], b[1]);
    ctx.lineTo(b[0] - 20 * Math.cos(ang - 0.4), b[1] - 20 * Math.sin(ang - 0.4));
    ctx.lineTo(b[0] - 20 * Math.cos(ang + 0.4), b[1] - 20 * Math.sin(ang + 0.4));
    ctx.closePath(); ctx.fill();
  }

  function dot(p, color, label, hollow, dy = -12) {
    const [x, y] = proj(p);
    ctx.beginPath(); ctx.arc(x, y, 9, 0, Math.PI * 2);
    ctx.fillStyle = hollow ? css('--card') : color; ctx.fill();
    ctx.lineWidth = 3; ctx.strokeStyle = color; ctx.stroke();
    if (label) { ctx.fillStyle = color; ctx.font = '600 28px -apple-system, sans-serif'; ctx.fillText(label, x + 14, y + dy); }
  }

  function label(p, text, color, dx, dy) {
    const [x, y] = proj(p);
    ctx.fillStyle = color; ctx.font = '600 26px -apple-system, sans-serif'; ctx.textAlign = dx < 0 ? 'right' : 'left'; ctx.fillText(text, x + dx, y + dy); ctx.textAlign = 'left';
  }

  function triangle(points, color) {
    const ps = points.map(proj);
    ctx.beginPath(); ctx.moveTo(ps[0][0], ps[0][1]);
    ps.slice(1).forEach(p => ctx.lineTo(p[0], p[1]));
    ctx.closePath(); ctx.fillStyle = color; ctx.globalAlpha = 0.1; ctx.fill(); ctx.globalAlpha = 1;
  }

  return { proj, curve, arrow, dot, label, triangle };
}
