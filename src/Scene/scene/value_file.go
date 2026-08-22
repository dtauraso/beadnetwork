
package scene

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
)

func LogPersistErr(label, path string, err error) {
	fmt.Fprintf(os.Stderr, "%s: persist %s: %v\n", label, path, err)
}

func ReadIfExists(path string, dst any) bool {
	raw, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	if err := decode(raw, dst); err != nil {
		LogPersistErr("valuefile: CORRUPT LEAF", path, err)
		return false
	}
	return true
}

func decode(raw []byte, dst any) error {
	switch p := dst.(type) {
	case *bool:
		if len(raw) != 1 {
			return fmt.Errorf("valuefile: bool leaf is %d bytes, want 1", len(raw))
		}
		*p = raw[0] != 0
		return nil
	case *int:
		v, err := readInt64(raw)
		*p = int(v)
		return err
	case *int32:
		v, err := readInt64(raw)
		*p = int32(v)
		return err
	case *int64:
		v, err := readInt64(raw)
		*p = v
		return err
	case *float64:
		if len(raw) != 8 {
			return fmt.Errorf("valuefile: float leaf is %d bytes, want 8", len(raw))
		}
		*p = math.Float64frombits(binary.LittleEndian.Uint64(raw))
		return nil
	case *string:
		*p = string(raw)
		return nil
	}
	return fmt.Errorf("valuefile: %T is not a leaf primitive pointer", dst)
}

func readInt64(raw []byte) (int64, error) {
	if len(raw) != 8 {
		return 0, fmt.Errorf("valuefile: integer leaf is %d bytes, want 8", len(raw))
	}
	return int64(binary.LittleEndian.Uint64(raw)), nil
}
