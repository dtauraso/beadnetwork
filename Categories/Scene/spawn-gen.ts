let lastGen = -1;
let restarted = false;

export function noteSpawnGen(gen: number): void {
  if (lastGen === -1) {
    lastGen = gen;
    return;
  }
  if (gen === lastGen) return;
  lastGen = gen;
  restarted = true;
}

export function takeSpawnRestarted(): boolean {
  if (!restarted) return false;
  restarted = false;
  return true;
}
