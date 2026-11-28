package Builder

import (
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Board"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Pills/Builder/StepStore"
	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
)

func (m Model) leavesUnder(p []int) [][]int {
	var out [][]int
	for _, l := range m.Leaves {
		if under(l, p) {
			out = append(out, l)
		}
	}
	return out
}

func (m *Model) runFolder(paths [][]int, now int64) []Links.Envelope {
	if len(paths) == 0 {
		return nil
	}
	m.Folder = &FolderRun{Paths: paths}
	return m.startRun(paths[0], now)
}

const OpFolderPlay = "folder-play"

type FolderRun struct {
	Paths [][]int
	I     int
}

func (m Model) folderPlayChip() *Board.Chip {
	c := Board.Chip{Label: "▶", Act: Act{Op: OpFolderPlay}}
	if f := m.Folder; f != nil {
		c.Label, c.Lit = "■ step "+itoa(f.I+1)+"/"+itoa(len(f.Paths)), true
	}
	return &c
}

func (m *Model) StartFolder(now int64) []Links.Envelope {
	return m.runFolder(m.leavesUnder(StepStore.LevelPath(m.Level)), now)
}

func (m *Model) NextFolderStep(now int64) ([]Links.Envelope, bool) {
	f := m.Folder
	if f == nil || m.Play != nil {
		return nil, false
	}
	if f.I++; f.I >= len(f.Paths) {
		m.Folder = nil
		return nil, false
	}
	return m.startRun(f.Paths[f.I], now), true
}

func (m *Model) StopPlay() (end []Links.Envelope, was bool) {
	folder := m.Folder != nil
	m.Folder = nil
	p := m.Play
	if p == nil {
		return nil, folder
	}
	m.Play = nil
	return m.playParts(p.Path, p.N, p.N), true
}

func (m *Model) PlayTick(now int64) ([]Links.Envelope, bool) {
	p := m.Play
	if p == nil || now < p.At {
		return nil, false
	}
	p.At = now + 1
	edits := m.playParts(p.Path, p.K, p.N)
	if p.K >= p.N {
		m.Play = nil
	}
	p.K++
	return edits, true
}
