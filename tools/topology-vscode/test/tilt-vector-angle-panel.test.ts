// tilt-vector-angle-panel.test.ts — pins TiltVectorAnglePanel's formatAngle deriving both
// the displayed INDEX and the fraction's denominator from the LIVE streamed lattice point
// count (Buffer/layout.go's LatticePoints), not the fixed compile-time
// CURVE_PARAM_TILT_VECTOR_ANGLE_STEP (π/12, a 24-point default) — task/pair-lattice-points.
//
// A streamed angle of π/2 is index 6 of 24 points (6π/12) or index 3 of 12 points (3π/6):
// same physical direction, different index and different denominator, because the step
// (2π/points) itself depends on the count.
import { describe, it, expect } from "vitest";
import { formatAngle } from "../src/webview/three/tilt-vector-angle-format";

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
