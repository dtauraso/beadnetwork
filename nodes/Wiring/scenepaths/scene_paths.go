package scenepaths

import "path/filepath"

func ViewFilePath(topologyPath, name string) string {
	return filepath.Join(topologyPath, "view", name)
}

func SelectionFilePath(anchorPath string) string {
	return ViewFilePath(anchorPath, "scene.json")
}

func CameraFilePath(topologyPath string) string {
	return ViewFilePath(topologyPath, "camera.json")
}

func OverlaysDirPath(topologyPath string) string {
	return ViewFilePath(topologyPath, "overlays")
}

func PanelsDirPath(topologyPath string) string {
	return ViewFilePath(topologyPath, "panels")
}

func SphereFilePath(topologyPath string) string {
	return ViewFilePath(topologyPath, "sphere.json")
}

func SpeedFilePath(topologyPath string) string {
	return ViewFilePath(topologyPath, "speed.json")
}

func LatticeFilePath(topologyPath string) string {
	return ViewFilePath(topologyPath, "lattice.json")
}
