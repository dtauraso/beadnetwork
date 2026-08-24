package main

//go:generate go run .

import (
	"github.com/dtauraso/beadnetwork/scripts/genpaths"
)

func main() {
	genpaths.SetName("Categories/Scene/ownercounts")
	_, srcRoot := genpaths.Roots()
	generateOwnerCounts(srcRoot)
}
