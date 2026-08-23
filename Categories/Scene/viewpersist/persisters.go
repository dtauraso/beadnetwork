package viewpersist

import (
	"github.com/dtauraso/wirefold/Categories/Chrome/Panels/Panel"
	"github.com/dtauraso/wirefold/Categories/Chrome/Pills/AngleDropdown"
	"github.com/dtauraso/wirefold/Categories/Overlay"
	"github.com/dtauraso/wirefold/Categories/Polar/polar"
	"github.com/dtauraso/wirefold/Categories/Scene/Camera"
	"github.com/dtauraso/wirefold/Categories/Scene/Scenes"
	"github.com/dtauraso/wirefold/Categories/Speed"
)

type Persisters struct {
	vp *Camera.ViewpointPersister

	overlays *Persister[Overlay.OverlayState]

	panels *Persister[Panel.PanelState]

	sphere *Persister[polar.SceneSphere]

	speed *Persister[float64]

	lattice *Persister[int32]
}

func (p *Persisters) ArmViewpoint(topologyPath string) *Camera.ViewpointPersister {
	vp := &Camera.ViewpointPersister{Path: Camera.BlockPath(topologyPath)}
	p.vp = vp
	return vp
}

func (p *Persisters) ArmEdit(topologyPath string) {
	p.overlays = &Persister[Overlay.OverlayState]{
		Path: topologyPath, Write: Overlay.WriteSceneOverlays, Tag: "scene_overlays_persist",
	}
	p.panels = &Persister[Panel.PanelState]{
		Path: topologyPath, Write: Panel.WriteScenePanels, Tag: "scene_panels_persist",
	}
	p.sphere = &Persister[polar.SceneSphere]{
		Path: topologyPath, Write: WriteSceneSphere, Tag: "scene_sphere_persist",
	}
	p.speed = &Persister[float64]{
		Path: Scenes.SpeedFilePath(topologyPath), Write: Speed.WriteSceneSpeed, Tag: "scene_speed_persist",
	}
	p.lattice = &Persister[int32]{
		Path: Scenes.LatticeFilePath(topologyPath), Write: AngleDropdown.WriteSceneLattice, Tag: "scene_lattice_persist",
	}
}

func (p *Persisters) Overlays() *Persister[Overlay.OverlayState] {
	return p.overlays
}

func (p *Persisters) Panels() *Persister[Panel.PanelState] { return p.panels }

func (p *Persisters) Sphere() *Persister[polar.SceneSphere] { return p.sphere }

func (p *Persisters) Speed() *Persister[float64] { return p.speed }

func (p *Persisters) Lattice() *Persister[int32] { return p.lattice }
