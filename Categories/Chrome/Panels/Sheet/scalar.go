package Sheet

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
)

func WriteScalar(path string, v int32) error {
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

func LoadScalar(path string, def, least int) int {
	raw, err := os.ReadFile(path)
	if err != nil {
		return def
	}
	if len(raw) != 4 {
		fmt.Fprintf(os.Stderr, "sheet scalar: %s is %d bytes, want 4\n", path, len(raw))
		return def
	}
	v := int(int32(binary.LittleEndian.Uint32(raw)))
	if v < least {
		return def
	}
	return v
}
