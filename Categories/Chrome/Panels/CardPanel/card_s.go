package CardPanel

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
)

func WriteCardS(path string, s int32) error {
	b := binary.LittleEndian.AppendUint32(nil, uint32(s))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func LoadCardS(path string) int {
	raw, err := os.ReadFile(path)
	if err != nil {
		return DefaultS
	}
	if len(raw) != 4 {
		fmt.Fprintf(os.Stderr, "card s: %s is %d bytes, want 4\n", path, len(raw))
		return DefaultS
	}
	s := int(int32(binary.LittleEndian.Uint32(raw)))
	if s < 1 {
		return DefaultS
	}
	return s
}
