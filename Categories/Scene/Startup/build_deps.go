package Startup

import "context"

type BuildDeps struct {
	LatticePoints int32

	ClaimLatticeIn func(name string) chan int32

	ClaimTiltEditIn func(name string) any

	ClaimSelfDriveGeom func(name string) any

	ClaimLinkIn func(name string) any

	Mesh any

	Seed any

	loadShared map[string]any
}

func (d BuildDeps) LoadShared(key string, mk func() any) any {
	v, ok := d.loadShared[key]
	if !ok {
		v = mk()
		d.loadShared[key] = v
	}
	return v
}

func (d BuildDeps) LatticePointsSeed() int32 { return d.LatticePoints }

func (d BuildDeps) LatticeChan(name string) chan int32 {
	if d.ClaimLatticeIn == nil {
		return nil
	}
	return d.ClaimLatticeIn(name)
}

func (d BuildDeps) TiltEditChan(name string) any {
	if d.ClaimTiltEditIn == nil {
		return nil
	}
	return d.ClaimTiltEditIn(name)
}

func (d BuildDeps) LinkInbox(name string) any {
	if d.ClaimLinkIn == nil {
		return nil
	}
	return d.ClaimLinkIn(name)
}

func (d BuildDeps) LinkMesh() any { return d.Mesh }

func (d BuildDeps) LinkSeed() any { return d.Seed }

func (d BuildDeps) SelfDriveGeom(name string) any {
	if d.ClaimSelfDriveGeom == nil {
		return nil
	}
	return d.ClaimSelfDriveGeom(name)
}

type BuiltNode interface {
	Update(ctx context.Context)
}
