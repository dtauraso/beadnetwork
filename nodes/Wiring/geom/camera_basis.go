package geom

// camera_basis.go — the three.js camera screen basis (mirrors BufferCamera.tsx lookAt +
// polar.ts cameraFrame), split from camera_angles.go by concern (see that file's header for
// the shared quarantine rationale).

// ---------------------------------------------------------------------------
// camera basis (mirrors BufferCamera.tsx lookAt + polar.ts cameraFrame)
// ---------------------------------------------------------------------------

// CamBasis is the three.js camera screen basis, reconstructed from the polar viewpoint
// (pos = pivot→camera dir; up = up-hint). It reproduces exactly what BufferCamera.tsx
// builds (cam.up = up; cam.lookAt(pivot)) and what polar.ts cameraFrame then reads off the
// quaternion:
//
//	pole (cam +Z, toward camera) = posWorld = AnglesToWorldOffset(1, pos)
//	refX (cam +X, screen right)  = normalize(upWorld × pole)   [three.js Matrix4.lookAt]
//	refY (cam +Y, screen up)     = pole × refX
type CamBasis struct {
	RefX vec3 // screen right
	RefY vec3 // screen up
	Pole vec3 // toward the camera (cam +Z)
}

func BasisFromViewpoint(pos, up Dir) CamBasis {
	pole := AnglesToWorldOffset(1, pos.Theta, pos.Phi) // unit
	upWorld := AnglesToWorldOffset(1, up.Theta, up.Phi).Normalize()
	refX := upWorld.Cross(pole).Normalize()
	refY := pole.Cross(refX)
	return CamBasis{RefX: refX, RefY: refY, Pole: pole}
}

// EyeOf is the camera world position for a viewpoint: pivot + r * posWorld.
func EyeOf(v Viewpoint) vec3 {
	return v.Pivot.Add(AnglesToWorldOffset(v.R, v.Pos.Theta, v.Pos.Phi))
}
