package Startup

import (
	"github.com/dtauraso/wirefold/Categories/Chrome/Panels/Panel"
	"github.com/dtauraso/wirefold/Categories/Chrome/Pills/AngleDropdown"
	"github.com/dtauraso/wirefold/Categories/Overlay"
	"github.com/dtauraso/wirefold/Categories/Polar/polar"
	SceneBuf "github.com/dtauraso/wirefold/Categories/Scene"
	"github.com/dtauraso/wirefold/Categories/Scene/Camera"
	"github.com/dtauraso/wirefold/Categories/Scene/Scenes"
	"github.com/dtauraso/wirefold/Categories/Scene/viewstate"
	"github.com/dtauraso/wirefold/Categories/Speed"
)

func armViewpoint(topologyPath string) *Camera.ViewpointPersister {
	return &Camera.ViewpointPersister{Path: Camera.BlockPath(topologyPath)}
}

func armEdit(ui *viewstate.UIState, topologyPath string) {
	overlays := &Persister[Overlay.OverlayState]{
		Path: topologyPath, Write: Overlay.WriteSceneOverlays, Tag: "scene_overlays_persist",
	}
	panels := &Persister[Panel.PanelState]{
		Path: topologyPath, Write: Panel.WriteScenePanels, Tag: "scene_panels_persist",
	}
	sphere := &Persister[polar.SceneSphere]{
		Path: topologyPath, Write: SceneBuf.WriteSceneSphere, Tag: "scene_sphere_persist",
	}
	speed := &Persister[float64]{
		Path: Scenes.SpeedFilePath(topologyPath), Write: Speed.WriteSceneSpeed, Tag: "scene_speed_persist",
	}
	lattice := &Persister[int32]{
		Path: Scenes.LatticeFilePath(topologyPath), Write: AngleDropdown.WriteSceneLattice, Tag: "scene_lattice_persist",
	}

	ui.PersistOverlays = overlays.Schedule
	ui.PersistPanels = panels.Schedule
	ui.PersistSphere = sphere.Schedule
	ui.PersistSpeed = speed.Schedule
	ui.PersistLattice = lattice.Schedule
}
