package Startup

import "context"

type BuildDeps struct {
	LatticePoints int32

	ClaimLatticeIn func(name string) chan int32

	ClaimTiltEditIn func(name string) any

	ClaimSelfDriveGeom func(name string) any

	ClaimCardEditIn func(name string) any

	CardS int

	CardM int
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

func (d BuildDeps) CardEditChan(name string) any {
	if d.ClaimCardEditIn == nil {
		return nil
	}
	return d.ClaimCardEditIn(name)
}

func (d BuildDeps) CardSSeed() int { return d.CardS }

func (d BuildDeps) CardMSeed() int { return d.CardM }

func (d BuildDeps) SelfDriveGeom(name string) any {
	if d.ClaimSelfDriveGeom == nil {
		return nil
	}
	return d.ClaimSelfDriveGeom(name)
}

type BuiltNode interface {
	Update(ctx context.Context)
}
