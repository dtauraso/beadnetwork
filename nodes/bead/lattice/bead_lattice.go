package lattice

import "github.com/dtauraso/wirefold/nodes/clock"

const PulseSpeedWuPerMs = 0.04

const PulseSpeedWuPerTick = PulseSpeedWuPerMs * clock.MsPerTick

const BeadRadius = 4.0

const BeadRingTubeRatio = 0.12

const BeadTorusOuterR = BeadRadius * (1 + BeadRingTubeRatio)

const BeadStepR = 2 * BeadTorusOuterR

const PulsesPerSlot = 14
