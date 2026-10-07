package Startup

import (
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/CardPanel"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Panel"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/SliderPanel"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Pills/AngleDropdown"
	SceneBuf "github.com/dtauraso/beadnetwork/Categories/Scene"
	"github.com/dtauraso/beadnetwork/Categories/Scene/Camera"
	"github.com/dtauraso/beadnetwork/Categories/Scene/Scenes"
	"github.com/dtauraso/beadnetwork/Categories/Scene/View"
	Flags "github.com/dtauraso/beadnetwork/Categories/Scene/View/Flags"
	"github.com/dtauraso/beadnetwork/Categories/Vectors/polar"
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
	cardS := &Persister[int32]{
		Path: Scenes.CardSFilePath(topologyPath), Write: CardPanel.WriteCardScalar, Tag: "card_s_persist",
	}
	cardM := &Persister[int32]{
		Path: Scenes.CardMFilePath(topologyPath), Write: CardPanel.WriteCardScalar, Tag: "card_m_persist",
	}

	ui.PersistLattice = lattice.Schedule
	ui.PersistCardS = cardS.Schedule
	ui.PersistCardM = cardM.Schedule
}
