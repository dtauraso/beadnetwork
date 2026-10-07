package CardPanel

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
)

func WriteCardScalar(path string, v int32) error {
	b := binary.LittleEndian.AppendUint32(nil, uint32(v))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func LoadCardScalar(path string, def int) int {
	raw, err := os.ReadFile(path)
	if err != nil {
		return def
	}
	if len(raw) != 4 {
		fmt.Fprintf(os.Stderr, "card scalar: %s is %d bytes, want 4\n", path, len(raw))
		return def
	}
	v := int(int32(binary.LittleEndian.Uint32(raw)))
	if v < 1 {
		return def
	}
	return v
}
