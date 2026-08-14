package bufschema

type bufLayoutPanel struct {
	Overlays uint8 `buf:"u8"`

	Node      uint8 `buf:"u8"`
	NodeShape uint8 `buf:"u8"`
	NodeState uint8 `buf:"u8"`
	NodeReach uint8 `buf:"u8"`
	NodePoles uint8 `buf:"u8"`
	NodeRules uint8 `buf:"u8"`

	Scene        uint8 `buf:"u8"`
	SceneGuides  uint8 `buf:"u8"`
	ScenePoles   uint8 `buf:"u8"`
	SceneVectors uint8 `buf:"u8"`
	SceneLabels  uint8 `buf:"u8"`
}
