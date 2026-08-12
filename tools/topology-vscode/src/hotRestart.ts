



import type { BuildResult } from "./goBuild";














export function shouldRestartAfterBuild(res: BuildResult): boolean {
  return res.ok && !res.busy;
}








export class TrailingDebouncer {
  private pending: ReturnType<typeof setTimeout> | undefined;

  constructor(private readonly delayMs: number) {}

  schedule(fn: () => void): void {
    if (this.pending) clearTimeout(this.pending);
    this.pending = setTimeout(() => {
      this.pending = undefined;
      fn();
    }, this.delayMs);
  }



  dispose(): void {
    if (this.pending) clearTimeout(this.pending);
    this.pending = undefined;
  }
}
