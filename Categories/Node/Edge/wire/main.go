package main

//go:generate go run .

import "github.com/dtauraso/wirefold/scripts/genpaths"

func main() {
	genpaths.SetName("Categories/Node/Edge/wire")
	_, srcRoot := genpaths.Roots()
	writeWireTS(srcRoot)
}
