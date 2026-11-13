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

type jsonRun struct {
	Sequence int          `json:"sequence"`
	Presses  []jsonButton `json:"presses"`
	Refused  []string     `json:"refused"`
	Before   jsonSequence `json:"before"`
	After    jsonSequence `json:"after"`
}

var (
	phiUp   = jsonButton{Links.KPhi, 1}
	thetaUp = jsonButton{Links.KTheta, 1}
	rUp     = jsonButton{Links.KR, 1}
)

var singles = []jsonButton{
	{Links.KPhi, 1}, {Links.KPhi, -1}, {Links.KTheta, 1}, {Links.KTheta, -1},
	{Links.KR, 1}, {Links.KR, -1}, {Links.KBoth, 1}, {Links.KBoth, -1},
}

var orders = [][]jsonButton{
	{phiUp, thetaUp, rUp}, {phiUp, rUp, thetaUp}, {thetaUp, phiUp, rUp},
	{thetaUp, rUp, phiUp}, {rUp, phiUp, thetaUp}, {rUp, thetaUp, phiUp},
}

type queued struct {
	to  string
	msg any
}

func run(seed Links.Seed, ids []string, presses []jsonButton) (Links.Seed, []string) {
	peers := map[string]*Links.Peer{}
	for _, id := range ids {
		peers[id] = Links.NewPeer(id, seed)
	}
	for round, b := range presses {
		var queue []queued
		for _, id := range ids {
			queue = append(queue, queued{id, Links.StepAll{Round: round + 1, K: b.K, Sign: b.Sign}})
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
	}
	return collect(seed, ids, peers)
}

func collect(seed Links.Seed, ids []string, peers map[string]*Links.Peer) (Links.Seed, []string) {
	after := Links.Seed{Frame: seed.Frame, Places: map[string]Links.Place{}, S: seed.S}
	refused := []string{}
	for _, id := range ids {
		p := peers[id]
		c := p.TakeChanges()
		if c.Refused != "" {
			refused = append(refused, c.Refused)
		}
		if _, had := seed.Places[id]; had || c.Place {
			after.Places[id] = *p.At
		}
		after.S = p.S
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

func dumpRuns(root string, seqs []int, constants polarindex.SceneConstants, lists [][]jsonButton) int {
	status := 0
	all := []jsonRun{}
	for _, k := range seqs {
		spec, err := loadSequence(root, k, constants)
		if err != nil {
			fmt.Fprintf(os.Stderr, "readscene: sequence %d: %v\n", k, err)
			status = 1
			continue
		}
		seed := Topology.LinkSeed(spec, sequenceS(root, k), 0)
		var ids []string
		for _, n := range spec.Nodes {
			ids = append(ids, n.ID)
		}
		before := sequenceJSON(root, k, spec, seed)
		for _, presses := range lists {
			after, refused := run(seed, ids, presses)
			all = append(all, jsonRun{Sequence: k, Presses: presses, Refused: refused, Before: before, After: sequenceJSON(root, k, spec, after)})
		}
	}
	if err := json.NewEncoder(os.Stdout).Encode(all); err != nil {
		fmt.Fprintf(os.Stderr, "readscene: %v\n", err)
		return 1
	}
	return status
}

func singlePresses() [][]jsonButton {
	var out [][]jsonButton
	for _, b := range singles {
		out = append(out, []jsonButton{b})
	}
	return out
}
