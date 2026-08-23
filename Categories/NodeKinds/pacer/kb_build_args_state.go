package pacer

func (a BuildArgs) StateSeed(key string, def int) int {
	if a.Data == nil || a.Data.State == nil {
		return def
	}
	if v, ok := a.Data.State[key]; ok {
		return v
	}
	return def
}
