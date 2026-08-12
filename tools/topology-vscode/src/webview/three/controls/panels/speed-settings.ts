
















export const SPEED_SETTINGS = [
  { speed: 0, label: "0" },
  { speed: 0.25, num: "1", den: "4" },
  { speed: 0.5, num: "1", den: "2" },
  { speed: 0.75, num: "3", den: "4" },
  { speed: 1, label: "1" },
  { speed: 2, label: "2" },
] as const;




export const settingKey = (s: (typeof SPEED_SETTINGS)[number]): string => String(s.speed);





export const DEFAULT_INDEX = SPEED_SETTINGS.findIndex((s) => s.speed === 1);




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
