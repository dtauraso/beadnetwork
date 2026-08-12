package wire

type Broadcast []*Out

func (outs Broadcast) PlaceDrivenAllAt(v int, dst []DriveItem, tick int64) []DriveItem {
	for _, o := range outs {
		if o == nil {
			continue
		}
		dst = append(dst, o.PlaceDrivenAt(v, tick))
	}
	return dst
}
