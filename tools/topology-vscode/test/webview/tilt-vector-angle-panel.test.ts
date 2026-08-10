// tilt-vector-angle-panel.test.ts — pins TiltVectorAnglePanel's formatAngle deriving both
// the displayed INDEX and the fraction's denominator from the LIVE streamed lattice point
// count (Buffer/layout.go's LatticePoints), not the fixed compile-time
// CURVE_PARAM_TILT_VECTOR_ANGLE_STEP (π/12, a 24-point default) — task/pair-lattice-points.
//
// A streamed angle of π/2 is index 6 of 24 points (6π/12) or index 3 of 12 points (3π/6):
// same physical direction, different index and different denominator, because the step
// (2π/points) itself depends on the count.
import { describe, it, expect } from "vitest";
import { formatAngle, widestAngle } from "../../src/webview/three/controls/tilt-vector-angle-format";

describe("TiltVectorAnglePanel formatAngle", () => {
  it("at 24 points, π/2 is index 6 shown as 6π/12", () => {
    expect(formatAngle(Math.PI / 2, 24)).toBe("6π/12");
  });

  it("at 12 points, the SAME radians (π/2) is index 3 shown as 3π/6", () => {
    expect(formatAngle(Math.PI / 2, 12)).toBe("3π/6");
  });

  it("zero radians is always \"0\", regardless of point count", () => {
    expect(formatAngle(0, 24)).toBe("0");
    expect(formatAngle(0, 12)).toBe("0");
  });

  it("a negative index carries its sign", () => {
    expect(formatAngle(-Math.PI / 2, 24)).toBe("-6π/12");
  });
});

// The readout reserves widestAngle(points) so the ▲/▼ beside it keep one position while θ
// steps. That only holds if no value formatAngle can produce is longer than the reservation.
describe("TiltVectorAnglePanel widestAngle", () => {
  it("is at least as long as every angle at that point count", () => {
    for (const points of [4, 12, 24, 64]) {
      const widest = widestAngle(points);
      const step = (2 * Math.PI) / points;
      for (let idx = -points; idx <= points; idx++) {
        expect(formatAngle(idx * step, points).length).toBeLessThanOrEqual(widest.length);
      }
    }
  });

  it("tracks the point count rather than a fixed character budget", () => {
    expect(widestAngle(24)).toBe("-24π/12");
    expect(widestAngle(12)).toBe("-12π/6");
  });
});
