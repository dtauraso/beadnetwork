package kindapi

import "github.com/dtauraso/wirefold/Scene/loadspec"

func (a BuildArgs) StateSeed(key string, def int) int {
	if a.data == nil || a.data.State == nil {
		return def
	}
	if v, ok := a.data.State[key]; ok {
		return v
	}
	return def
}

func (a BuildArgs) Data() *loadspec.NodeData { return a.data }
