package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
	"github.com/dtauraso/beadnetwork/Categories/Scene/Topology"
	"github.com/dtauraso/beadnetwork/Categories/Vectors/polarindex"
)

type jsonButton struct {
	K    string `json:"k"`
	Sign int    `json:"sign"`
}

type jsonStep struct {
	Sequence int          `json:"sequence"`
	Button   jsonButton   `json:"button"`
	Refused  []string     `json:"refused"`
	Before   jsonSequence `json:"before"`
	After    jsonSequence `json:"after"`
}

var buttons = []jsonButton{
	{Links.KPhi, 1}, {Links.KPhi, -1}, {Links.KTheta, 1}, {Links.KTheta, -1},
	{Links.KR, 1}, {Links.KR, -1}, {Links.KBoth, 1}, {Links.KBoth, -1},
}

type queued struct {
	to  string
	msg any
}

func press(seed Links.Seed, ids []string, b jsonButton) (Links.Seed, []string) {
	peers := map[string]*Links.Peer{}
	var queue []queued
	for _, id := range ids {
		peers[id] = Links.NewPeer(id, seed)
		queue = append(queue, queued{id, Links.StepAll{Round: 1, K: b.K, Sign: b.Sign}})
	}
	for len(queue) > 0 {
		q := queue[0]
		queue = queue[1:]
		p, ok := peers[q.to]
		if !ok {
			continue
		}
		for _, e := range p.Handle(q.msg) {
			if e.To != Links.All {
				queue = append(queue, queued{e.To, e.Msg})
				continue
			}
			for _, id := range ids {
				if id != q.to {
					queue = append(queue, queued{id, e.Msg})
				}
			}
		}
	}
	after := Links.Seed{Frame: seed.Frame, Places: map[string]Links.Place{}, Scales: map[string]int{}}
	var refused []string
	for _, id := range ids {
		p := peers[id]
		c := p.TakeChanges()
		if c.Refused != "" {
			refused = append(refused, c.Refused)
		}
		if _, had := seed.Places[id]; had || c.Place {
			after.Places[id] = p.At
		}
		after.Scales[id] = p.S
	}
	for _, l := range seed.Links {
		if p, ok := peers[l.To]; ok {
			for _, m := range p.Links() {
				if m.From == l.From {
					l = m
				}
			}
		}
		after.Links = append(after.Links, l)
	}
	return after, refused
}

func dumpSteps(root string, seqs []int, constants polarindex.SceneConstants, sceneS int) int {
	status := 0
	all := []jsonStep{}
	for _, k := range seqs {
		spec, err := loadSequence(root, k, constants)
		if err != nil {
			fmt.Fprintf(os.Stderr, "readscene: sequence %d: %v\n", k, err)
			status = 1
			continue
		}
		seed := Topology.LinkSeed(spec, sceneS, 0)
		var ids []string
		for _, n := range spec.Nodes {
			ids = append(ids, n.ID)
		}
		before := sequenceJSON(root, k, spec, seed)
		for _, b := range buttons {
			after, refused := press(seed, ids, b)
			if refused == nil {
				refused = []string{}
			}
			all = append(all, jsonStep{Sequence: k, Button: b, Refused: refused, Before: before, After: sequenceJSON(root, k, spec, after)})
		}
	}
	enc := json.NewEncoder(os.Stdout)
	if err := enc.Encode(all); err != nil {
		fmt.Fprintf(os.Stderr, "readscene: %v\n", err)
		return 1
	}
	return status
}
