package nodeactor

func (c *nodeClocks) Tick() int64 { return c.clk.Tick() }

func (c *nodeClocks) CopyClockSrc() {
	if c.clockSrc != nil {
		c.clk = c.clockSrc.Copy()
	}
}
