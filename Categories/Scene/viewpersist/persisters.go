package viewpersist

import (
	"github.com/dtauraso/wirefold/Categories/Camera"
	"github.com/dtauraso/wirefold/Categories/Chrome/Panels/Panel"
	"github.com/dtauraso/wirefold/Categories/Chrome/Pills/AngleDropdown"
	"github.com/dtauraso/wirefold/Categories/Overlay"
	"github.com/dtauraso/wirefold/Categories/Polar/polar"
	"github.com/dtauraso/wirefold/Categories/Scene/scenepaths"
	"github.com/dtauraso/wirefold/Categories/Scene/scenepersist"
)

type Persisters struct {
	vp *Camera.ViewpointPersister

	overlays *scenepersist.Persister[Overlay.OverlayState]

	panels *scenepersist.Persister[Panel.PanelState]

	sphere *scenepersist.Persister[polar.SceneSphere]

	speed *scenepersist.Persister[float64]

	lattice *scenepersist.Persister[int32]
}

func (p *Persisters) ArmViewpoint(topologyPath string) *Camera.ViewpointPersister {
	vp := &Camera.ViewpointPersister{Path: Camera.BlockPath(topologyPath)}
	p.vp = vp
	return vp
}

func (p *Persisters) ArmEdit(topologyPath string) {
	p.overlays = &scenepersist.Persister[Overlay.OverlayState]{
		Path: topologyPath, Write: Overlay.WriteSceneOverlays, Tag: "scene_overlays_persist",
	}
	p.panels = &scenepersist.Persister[Panel.PanelState]{
		Path: topologyPath, Write: Panel.WriteScenePanels, Tag: "scene_panels_persist",
	}
	p.sphere = &scenepersist.Persister[polar.SceneSphere]{
		Path: topologyPath, Write: scenepersist.WriteSceneSphere, Tag: "scene_sphere_persist",
	}
	p.speed = &scenepersist.Persister[float64]{
		Path: scenepaths.SpeedFilePath(topologyPath), Write: scenepersist.WriteSceneSpeed, Tag: "scene_speed_persist",
	}
	p.lattice = &scenepersist.Persister[int32]{
		Path: scenepaths.LatticeFilePath(topologyPath), Write: AngleDropdown.WriteSceneLattice, Tag: "scene_lattice_persist",
	}
}

func (p *Persisters) Overlays() *scenepersist.Persister[Overlay.OverlayState] {
	return p.overlays
}

func (p *Persisters) Panels() *scenepersist.Persister[Panel.PanelState] { return p.panels }

func (p *Persisters) Sphere() *scenepersist.Persister[polar.SceneSphere] { return p.sphere }

func (p *Persisters) Speed() *scenepersist.Persister[float64] { return p.speed }

func (p *Persisters) Lattice() *scenepersist.Persister[int32] { return p.lattice }
