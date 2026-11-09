package tiltring

func Mod(x, tau int32) int32 { return (x%tau + tau) % tau }

func Bottom(top, tau int32) int32 { return Mod(top+tau/2, tau) }

func Sent(top, tau int32) int32 { return Mod(top+tau/4, tau) }
