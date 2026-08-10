// geom_bridge.go — Wiring's own local aliases for the self-contained polar/camera geometry
// cluster, which lives in nodes/Wiring/geom now (polar.go, spherical.go, viewpoint.go,
// gesture_camera.go, sphere_layout.go). Mirrors the alias curve_params.go already keeps for
// vec3/wireSegment (nodes/wire/geometry.go): the types and funcs below are DEFINED in geom,
// Wiring keeps these short names so every existing call site (dir{...}, polar{...},
// polar2cart(...), ...) reads unchanged. Pure code motion — no behaviour change. Only the
// symbols Wiring's OWN files actually still reference are kept here; geom's other exports
// (Rot, CamBasis, PolarDir, WrapPi, AzimuthFrom, ...) are used only inside geom itself and
// are called there as geom.X, not through this bridge.
package Wiring

import (
	geom "github.com/dtauraso/wirefold/nodes/Wiring/geom"
)

// Types.
type polar = geom.Polar
type dir = geom.Dir
type viewpoint = geom.Viewpoint
type sphereEdge = geom.SphereEdge
type sceneSphere = geom.SceneSphere

// Consts.
const gestureZoomBase = geom.GestureZoomBase
const rotSmoothAlpha = geom.RotSmoothAlpha

// Funcs (package-level function VALUES, not wrappers — identical to calling geom.X
// directly, kept under the old short name so call sites need no rewrite).
var (
	polar2cart            = geom.Polar2cart
	inwardPole            = geom.InwardPole
	cart2polar            = geom.Cart2polar
	polarDist             = geom.PolarDist
	clamp                 = geom.Clamp
	angularDistance       = geom.AngularDistance
	anglesToWorldOffset   = geom.AnglesToWorldOffset
	worldDirToAngles      = geom.WorldDirToAngles
	basisFromViewpoint    = geom.BasisFromViewpoint
	eyeOf                 = geom.EyeOf
	screenToPolar         = geom.ScreenToPolar
	toWorldDir            = geom.ToWorldDir
	planeSlide            = geom.PlaneSlide
	deltaToPolar          = geom.DeltaToPolar
	panDisplacementPolar  = geom.PanDisplacementPolar
	focusAhead            = geom.FocusAhead
	contentSphereOf       = geom.ContentSphereOf
	regionFocus           = geom.RegionFocus
	fitDistanceGo         = geom.FitDistanceGo
	homeFitPose           = geom.HomeFitPose
	projectNDC            = geom.ProjectNDC
	rayDirThroughNDC      = geom.RayDirThroughNDC
	contentFitSceneSphere = geom.ContentFitSceneSphere
)
