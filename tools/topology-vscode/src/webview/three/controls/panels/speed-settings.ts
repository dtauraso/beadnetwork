// speed-settings.ts — the fixed playback-speed table SpeedSlider's <input> indexes into,
// and the pure lookup from a decoded float32 speed back to its table position. No
// react/vscode-api dependency, split the same way tilt-vector-angle-format.ts is split from
// its panel.
//
// SPEED_SETTINGS — the six settings the slider's raw <input> value (0..5) indexes into.
// The slider position is an INDEX, never the multiplier itself: `speed` is what crosses the
// wire, `label` is how that position reads on screen.
//
// One array of pairs rather than a value table beside a label table: two parallel arrays
// can lose step with each other (a setting added to one and not the other), and nothing
// would catch it — the labels would silently name the wrong speeds from that index on.
//
// The quarters read as fractions rather than decimals because the table IS quarters of the
// clock (nodes/wire.MsPerTick is divisible by 4 so they divide exactly —
// clock_ms_per_tick_quarters_test.go pins that); a fraction says that where "0.25" reads as
// an arbitrary decimal.
export const SPEED_SETTINGS = [
  { speed: 0, label: "0" },
  { speed: 0.25, num: "1", den: "4" },
  { speed: 0.5, num: "1", den: "2" },
  { speed: 0.75, num: "3", den: "4" },
  { speed: 1, label: "1" },
  { speed: 2, label: "2" },
] as const;

// settingKey identifies a setting for React's list reconciliation — the speed itself, which
// is unique across the table by construction (a label is not: "1" appears as both a whole
// number and a numerator).
export const settingKey = (s: (typeof SPEED_SETTINGS)[number]): string => String(s.speed);

// DEFAULT_INDEX is the position holding multiplier 1 — what the slider shows until the
// first snapshot decodes, matching Go's own defaultPlaybackSpeed fallback for a missing or
// malformed speed.json. Found by value rather than written as a literal 4, so reordering
// or inserting a setting cannot leave it pointing at the wrong one.
export const DEFAULT_INDEX = SPEED_SETTINGS.findIndex((s) => s.speed === 1);

// closestSettingIndex finds the SPEED_SETTINGS position whose speed is nearest a decoded
// float32 (a float32 round-trip of e.g. 0.25 is exact, but this guards against any future
// non-table value reaching the buffer, rather than an exact lookup silently missing).
export function closestSettingIndex(speed: number): number {
  let best = 0;
  let bestDiff = Infinity;
  SPEED_SETTINGS.forEach((setting, i) => {
    const diff = Math.abs(setting.speed - speed);
    if (diff < bestDiff) {
      bestDiff = diff;
      best = i;
    }
  });
  return best;
}
