package Speed

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"path/filepath"
)

// What a scene remembers about its speed, and the two conversions that go with
// it. The speed is the speed's business, so the file it is kept in and the
// slider number it shows as live here rather than with the machinery that
// happens to write files.

const DefaultPlaybackSpeed = 1.0

// SliderNum is the integer the slider shows for a speed.
func SliderNum(userSpeed float64) int64 {
	return int64(math.Round(userSpeed * SpeedNumScale))
}

// EffectiveClockSpeed is the speed the clock actually runs at: what the user
// asked for, divided by the scene's own divisor.
func EffectiveClockSpeed(userSpeed, divisor float64) float64 {
	if divisor <= 0 {
		return userSpeed
	}
	return userSpeed / divisor
}

func WriteSceneSpeed(speedPath string, speed float64) error {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, math.Float64bits(speed))
	if err := os.MkdirAll(filepath.Dir(speedPath), 0o755); err != nil {
		return err
	}
	tmp := speedPath + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, speedPath)
}

func LoadSceneSpeed(speedPath string) (float64, bool) {
	raw, err := os.ReadFile(speedPath)
	if err != nil {
		return DefaultPlaybackSpeed, false
	}
	if len(raw) != 8 {
		fmt.Fprintf(os.Stderr, "scene speed: %s is %d bytes, want 8\n", speedPath, len(raw))
		return DefaultPlaybackSpeed, false
	}
	return math.Float64frombits(binary.LittleEndian.Uint64(raw)), true
}
