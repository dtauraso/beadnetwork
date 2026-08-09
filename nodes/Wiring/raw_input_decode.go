// raw_input_decode.go — decode of the ONE record kind that is not an addressed edit.
//
// A raw-input record is a flat run of fixed-width fields (pointer state + the stateless
// raycast hit) in the exact order INPUT_LAYOUT_FINGERPRINT pins, so it reads as a straight
// sequence rather than a dispatch — which is why it sits apart from input_codec.go's
// per-kind edit decoding. Go's gesture FSM decides what the event MEANS; this file only
// says where each number is.

package Wiring

func decodeRawInput(r *recReader) (rawInputMsg, bool) {
	var ev rawInputMsg
	var e error
	f := func() float64 {
		v, err := r.f64()
		if err != nil && e == nil {
			e = err
		}
		return v
	}
	i := func() int {
		v, err := r.i32()
		if err != nil && e == nil {
			e = err
		}
		return int(v)
	}
	b := func() bool {
		v, err := r.boolByte()
		if err != nil && e == nil {
			e = err
		}
		return v
	}
	u := func() byte {
		v, err := r.u8()
		if err != nil && e == nil {
			e = err
		}
		return v
	}

	ev.Kind = enumAt(inEventKinds, u())
	ev.X = f()
	ev.Y = f()
	ev.RectLeft = f()
	ev.RectTop = f()
	ev.RectWidth = f()
	ev.RectHeight = f()
	ev.Button = i()
	ev.Ctrl = b()
	ev.Shift = b()
	ev.Alt = b()
	ev.Meta = b()
	ev.DeltaX = f()
	ev.DeltaY = f()
	ev.Fov = f()
	ev.Hit.Kind = enumAt(inHitKinds, u())
	ev.Hit.IsInput = b()
	ev.Hit.NodeRow = i()
	ev.Hit.PortRow = i()
	ev.Hit.EdgeRow = i()
	if e != nil || ev.Kind == "" || ev.Hit.Kind == "" {
		return ev, false
	}
	return ev, true
}
