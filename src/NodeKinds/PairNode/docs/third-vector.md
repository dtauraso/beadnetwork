# PairNode — third vector (received direction)

[← BEHAVIOR.md](BEHAVIOR.md)

Alongside its own tilt vector and the coplanar normal, this node draws a THIRD arrow:
the direction that last ARRIVED on its vector channel (`ReceivedThetaIdx`/
streamed as the buffer's `ReceivedVectorLen`/`ReceivedVectorTheta` columns,
`src/schema/buffer-layout/layout_version.go`). It:

- Persists indefinitely once set — it is NOT cleared when the straightening exchange
  settles (i.e. the arrival lands on this node's own top, so nothing steps and nothing is
  sent). An arrival is
  recorded even when it moves nothing: the last direction this node was sent is what it is
  still holding, and blanking the arrow when the pair comes to rest would erase the state
  it came to rest in.
- Is REPLACED, never accumulated, by the next arrival.
- Is cleared ONLY by a reset — this node's own (`TiltEditIn`'s `Reset`) or a Reset
  marker received on the channel — both zero `ReceivedSet`, and `ReceivedVectorLen`
  streams 0 in that state.
- Is distinguishable from "received (0,0)" (world +y): `ReceivedVectorLen` is 0 only
  when nothing has been received yet or a reset cleared it; an actually-received (0,0)
  direction still streams a non-zero length (this node's own radius, same as
  `TopTiltVectorLen`).
- Draws in its OWN colour (`RECEIVED_VECTOR_COLOR`, `TiltVectors.tsx`), distinct from
  the tilt vector/coplanar normal's shared magenta.
