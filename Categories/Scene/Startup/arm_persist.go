package Startup

import (
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Panel"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Pills/AngleDropdown"
	Flags "github.com/dtauraso/beadnetwork/Categories/Scene/View/Flags"
	"github.com/dtauraso/beadnetwork/Categories/Vectors/polar"
	SceneBuf "github.com/dtauraso/beadnetwork/Categories/Scene"
	"github.com/dtauraso/beadnetwork/Categories/Scene/Camera"
	"github.com/dtauraso/beadnetwork/Categories/Scene/Scenes"
	"github.com/dtauraso/beadnetwork/Categories/Scene/View"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/SliderPanel"
)

func armViewpoint(topologyPath string) *Camera.ViewpointPersister {
	return &Camera.ViewpointPersister{Path: Camera.BlockPath(topologyPath)}
}

func armEdit(ui *View.UIState, topologyPath string) {
	overlays := &Persister[Flags.OverlayState]{
		Path: topologyPath, Write: Flags.WriteSceneOverlays, Tag: "scene_overlays_persist",
	}
	panels := &Persister[Panel.PanelState]{
		Path: topologyPath, Write: Panel.WriteScenePanels, Tag: "scene_panels_persist",
	}
	sphere := &Persister[polar.SceneSphere]{
		Path: topologyPath, Write: SceneBuf.WriteSceneSphere, Tag: "scene_sphere_persist",
	}
	speed := &Persister[float64]{
		Path: Scenes.SpeedFilePath(topologyPath), Write: SliderPanel.WriteSceneSpeed, Tag: "scene_speed_persist",
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
