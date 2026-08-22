package Codec

func decodeRawInput(r *Reader) (RawInputMsg, bool) {
	var ev RawInputMsg
	var e error
	f := func() float64 {
		v, err := r.F64()
		if err != nil && e == nil {
			e = err
		}
		return v
	}
	i := func() int {
		v, err := r.I32()
		if err != nil && e == nil {
			e = err
		}
		return int(v)
	}
	b := func() bool {
		v, err := r.BoolByte()
		if err != nil && e == nil {
			e = err
		}
		return v
	}
	u := func() byte {
		v, err := r.U8()
		if err != nil && e == nil {
			e = err
		}
		return v
	}

	ev.Kind = EnumAt(InEventKinds, u())
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
	ev.Hit.Kind = EnumAt(InHitKinds, u())
	ev.Hit.IsInput = b()
	ev.Hit.NodeRow = i()
	ev.Hit.PortRow = i()
	ev.Hit.EdgeRow = i()
	if key, err := r.Str(); err == nil {
		ev.Key = key
	} else if e == nil {
		e = err
	}
	if e != nil || ev.Kind == "" || ev.Hit.Kind == "" {
		return ev, false
	}
	return ev, true
}
