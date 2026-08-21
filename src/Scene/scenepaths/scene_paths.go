package scenepaths

import "path/filepath"

func ViewFilePath(topologyPath, name string) string {
	return filepath.Join(topologyPath, "view", name)
}

func InputFilePath(topologyPath string) string {
	return ViewFilePath(topologyPath, filepath.Join("input", "current.bin"))
}

func SelectionFilePath(anchorPath string) string {
	return ViewFilePath(anchorPath, filepath.Join("scene", "selected.bin"))
}

func CameraDirPath(topologyPath string) string {
	return ViewFilePath(topologyPath, "camera")
}

func SphereDirPath(topologyPath string) string {
	return ViewFilePath(topologyPath, "sphere")
}

func SpeedFilePath(topologyPath string) string {
	return ViewFilePath(topologyPath, "speed.bin")
}

func LatticeFilePath(topologyPath string) string {
	return ViewFilePath(topologyPath, "lattice.bin")
}
