package Camera

type CamBasis struct {
	RefX Vec3
	RefY Vec3
	Pole Vec3
}

func BasisFromViewpoint(pos, up Dir) CamBasis {
	pole := AnglesToWorldOffset(1, pos.Phi, pos.Theta)
	upWorld := AnglesToWorldOffset(1, up.Phi, up.Theta).Normalize()
	refX := upWorld.Cross(pole).Normalize()
	refY := pole.Cross(refX)
	return CamBasis{RefX: refX, RefY: refY, Pole: pole}
}

func EyeOf(v Viewpoint) Vec3 {
	return v.Pivot.Add(AnglesToWorldOffset(v.R, v.Pos.Phi, v.Pos.Theta))
}
