package main

//go:generate go run .

func main() {
	genName = "Categories/Input/gen"
	repoRoot, _ := roots()

	copyRecordReaders(repoRoot)
	copyTSWireVocabulary(repoRoot)
}
