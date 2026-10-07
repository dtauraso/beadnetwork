package Scenes

import "path/filepath"

func ViewFilePath(topologyPath, name string) string {
	return filepath.Join(topologyPath, "view", name)
}

func CardSFilePath(topologyPath string) string {
	return ViewFilePath(topologyPath, "card-s.bin")
}

func InputDirPath(topologyPath string) string {
	return ViewFilePath(topologyPath, "input")
}

func SelectionFilePath(anchorPath string) string {
	return ViewFilePath(anchorPath, filepath.Join("scene", "selected.bin"))
}

func SpeedFilePath(topologyPath string) string {
	return ViewFilePath(topologyPath, "speed.bin")
}

func LatticeFilePath(topologyPath string) string {
	return ViewFilePath(topologyPath, "lattice.bin")
}
