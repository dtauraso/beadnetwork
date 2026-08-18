export function subscribeFrame(fn: () => void): () => void {
  let alive = true;
  const tick = () => {
    if (!alive) return;
    fn();
    requestAnimationFrame(tick);
  };
  requestAnimationFrame(tick);
  return () => { alive = false; };
}
