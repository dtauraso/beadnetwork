let lastGen = -1;
let restarted = false;
const listeners: (() => void)[] = [];

export function onSpawnRestart(fn: () => void): void {
  listeners.push(fn);
}

export function noteSpawnGen(gen: number): void {
  if (lastGen === -1) {
    lastGen = gen;
    return;
  }
  if (gen === lastGen) return;
  lastGen = gen;
  restarted = true;
  for (const fn of listeners) fn();
}

export function takeSpawnRestarted(): boolean {
  if (!restarted) return false;
  restarted = false;
  return true;
}
