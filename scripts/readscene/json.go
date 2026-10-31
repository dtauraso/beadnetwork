package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/dtauraso/beadnetwork/Categories/Scene/Scenes"
	"github.com/dtauraso/beadnetwork/Categories/Scene/Topology"
	"github.com/dtauraso/beadnetwork/Categories/Vectors/polarindex"
)

type jsonAngles struct {
	Phi   float64 `json:"phi"`
	Theta float64 `json:"theta"`
	R     float64 `json:"r"`
}

type jsonTicks struct {
	Phi   int `json:"phi"`
	Theta int `json:"theta"`
	R     int `json:"r"`
}

type jsonFrame struct {
	Rho           float64 `json:"rho"`
	ConstantR     float64 `json:"constantR"`
	MaxIndexPhi   int     `json:"maxIndexPhi"`
	MaxIndexTheta int     `json:"maxIndexTheta"`
}

type jsonNode struct {
	ID    string      `json:"id"`
	S     int         `json:"s"`
	Place *jsonAngles `json:"place"`
}

type jsonLink struct {
	From      string     `json:"from"`
	To        string     `json:"to"`
	Start     jsonAngles `json:"start"`
	Delta     jsonTicks  `json:"delta"`
	Increment jsonTicks  `json:"increment"`
	Picked    bool       `json:"picked"`
}

type jsonSequence struct {
	Sequence int        `json:"sequence"`
	Name     string     `json:"name"`
	Frame    jsonFrame  `json:"frame"`
	Nodes    []jsonNode `json:"nodes"`
	Links    []jsonLink `json:"links"`
}

func dumpJSON(root string, seqs []int, constants polarindex.SceneConstants, sceneS int) int {
	status := 0
	all := []jsonSequence{}
	for _, k := range seqs {
		spec, err := loadSequence(root, k, constants)
		if err != nil {
			fmt.Fprintf(os.Stderr, "readscene: sequence %d: %v\n", k, err)
			status = 1
			continue
		}
		seed := Topology.LinkSeed(spec, sceneS, 0)
		f := seed.Frame
		js := jsonSequence{Sequence: k, Name: Scenes.SequenceName(root, k),
			Frame: jsonFrame{Rho: f.Rho, ConstantR: f.Cr, MaxIndexPhi: f.Tphi, MaxIndexTheta: f.Ttheta},
			Nodes: []jsonNode{}, Links: []jsonLink{}}
		for _, n := range spec.Nodes {
			jn := jsonNode{ID: n.ID, S: seed.Scales[n.ID]}
			if p, ok := seed.Places[n.ID]; ok {
				jn.Place = &jsonAngles{Phi: p.Phi, Theta: p.Theta, R: p.R}
			}
			js.Nodes = append(js.Nodes, jn)
		}
		for _, l := range spec.Links {
			js.Links = append(js.Links, jsonLink{From: l.From, To: l.To,
				Start:     jsonAngles{Phi: l.Start.Phi, Theta: l.Start.Theta, R: l.Start.R},
				Delta:     jsonTicks{Phi: l.Delta.Phi, Theta: l.Delta.Theta, R: l.Delta.R},
				Increment: jsonTicks{Phi: l.Increment.Phi, Theta: l.Increment.Theta, R: l.Increment.R},
				Picked:    l.Picked})
		}
		all = append(all, js)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(all); err != nil {
		fmt.Fprintf(os.Stderr, "readscene: %v\n", err)
		return 1
	}
	return status
}
