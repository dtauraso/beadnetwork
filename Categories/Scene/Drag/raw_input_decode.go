package Drag

func decodeRawInputFrom(r *Reader) (RawInputMsg, bool) {
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

	ev.Kind = enumAt(EventKinds, u())
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
	ev.Hit.Kind = enumAt(HitKinds, u())
	ev.Hit.IsInput = b()
	ev.Hit.OnRim = b()
	ev.Hit.NodeRow = i()
	ev.Hit.PortRow = i()
	ev.Hit.EdgeRow = i()
	if key, err := r.Str(); err == nil {
		ev.Key = key
	} else if e == nil {
		e = err
	}
	ev.Hit.Point = Vec3{X: f(), Y: f(), Z: f()}
	ev.Ball = Vec3{X: f(), Y: f(), Z: f()}
	ev.BallPrev = Vec3{X: f(), Y: f(), Z: f()}
	if e != nil || ev.Kind == "" || ev.Hit.Kind == "" {
		return ev, false
	}
	return ev, true
}
