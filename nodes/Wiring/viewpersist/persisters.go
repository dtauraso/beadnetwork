package viewpersist

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/camerapersist"
	"github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"
	"github.com/dtauraso/wirefold/nodes/Wiring/scenepaths"
	"github.com/dtauraso/wirefold/nodes/Wiring/scenepersist"
	"github.com/dtauraso/wirefold/nodes/Wiring/viewstate"
)

type Persisters struct {
	vp *camerapersist.ViewpointPersister

	overlays *scenepersist.Persister[viewstate.OverlayState]

	panels *scenepersist.Persister[viewstate.PanelState]

	sphere *scenepersist.Persister[polar.SceneSphere]

	speed *scenepersist.Persister[float64]

	lattice *scenepersist.Persister[int32]
}

func (p *Persisters) ArmViewpoint(topologyPath string) *camerapersist.ViewpointPersister {
	vp := &camerapersist.ViewpointPersister{Path: scenepaths.CameraFilePath(topologyPath)}
	p.vp = vp
	return vp
}

func (p *Persisters) ArmEdit(topologyPath string) {
	p.overlays = &scenepersist.Persister[viewstate.OverlayState]{
		Path: scenepaths.OverlaysFilePath(topologyPath), Write: scenepersist.WriteSceneOverlays, Tag: "scene_overlays_persist",
	}
	p.panels = &scenepersist.Persister[viewstate.PanelState]{
		Path: scenepaths.PanelsFilePath(topologyPath), Write: scenepersist.WriteScenePanels, Tag: "scene_panels_persist",
	}
	p.sphere = &scenepersist.Persister[polar.SceneSphere]{
		Path: scenepaths.SphereFilePath(topologyPath), Write: scenepersist.WriteSceneSphere, Tag: "scene_sphere_persist",
	}
	p.speed = &scenepersist.Persister[float64]{
		Path: scenepaths.SpeedFilePath(topologyPath), Write: scenepersist.WriteSceneSpeed, Tag: "scene_speed_persist",
	}
	p.lattice = &scenepersist.Persister[int32]{
		Path: scenepaths.LatticeFilePath(topologyPath), Write: scenepersist.WriteSceneLattice, Tag: "scene_lattice_persist",
	}
}

func (p *Persisters) Overlays() *scenepersist.Persister[viewstate.OverlayState] { return p.overlays }

func (p *Persisters) Panels() *scenepersist.Persister[viewstate.PanelState] { return p.panels }

func (p *Persisters) Sphere() *scenepersist.Persister[polar.SceneSphere] { return p.sphere }

func (p *Persisters) Speed() *scenepersist.Persister[float64] { return p.speed }

func (p *Persisters) Lattice() *scenepersist.Persister[int32] { return p.lattice }
