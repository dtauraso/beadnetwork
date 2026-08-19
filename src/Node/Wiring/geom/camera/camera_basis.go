package camera

type CamBasis struct {
	RefX vec3
	RefY vec3
	Pole vec3
}

func BasisFromViewpoint(pos, up Dir) CamBasis {
	pole := AnglesToWorldOffset(1, pos.Phi, pos.Theta)
	upWorld := AnglesToWorldOffset(1, up.Phi, up.Theta).Normalize()
	refX := upWorld.Cross(pole).Normalize()
	refY := pole.Cross(refX)
	return CamBasis{RefX: refX, RefY: refY, Pole: pole}
}

func EyeOf(v Viewpoint) vec3 {
	return v.Pivot.Add(AnglesToWorldOffset(v.R, v.Pos.Phi, v.Pos.Theta))
}
