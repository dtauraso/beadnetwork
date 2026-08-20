package jsonpersist

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func SafeTreePathComponent(s string) bool {
	return s != "" && s != "." && s != ".." && !filepath.IsAbs(s) &&
		!strings.ContainsRune(s, '/') && !strings.ContainsRune(s, '\\') &&
		s == filepath.Base(s)
}

const atomicWriteTmpSuffix = ".tmp"

func LogPersistErr(label, path string, err error) {
	fmt.Fprintf(os.Stderr, "%s: write %s: %v\n", label, path, err)
}

func ReadJSONBestEffort(path string, v any) {
	ReadJSONIfExists(path, v)
}

func ReadJSONIfExists(path string, v any) bool {
	raw, err := os.ReadFile(path)
	if err != nil || len(raw) == 0 {
		return false
	}
	return json.Unmarshal(raw, v) == nil
}

func ReadModifyWriteJSON(path string, mutate func(map[string]any)) error {
	m := map[string]any{}
	raw, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("ReadModifyWriteJSON: read %s: %w", path, err)
	}
	if err == nil && len(raw) > 0 {
		if err := json.Unmarshal(raw, &m); err != nil {
			return fmt.Errorf(
				"ReadModifyWriteJSON: %s exists but does not parse as JSON (%w) — refusing to "+
					"overwrite it, since that would discard every key this write does not set",
				path, err)
		}
	}
	mutate(m)
	return WriteJSONAtomic(path, m)
}

func RemoveIfPresent(path string) error {
	err := os.Remove(path)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func WriteJSONAtomicIfChanged(path string, v any) error {
	out, err := json.Marshal(v)
	if err != nil {
		return err
	}
	if prev, err := os.ReadFile(path); err == nil && bytes.Equal(prev, out) {
		return nil
	}
	return writeBytesAtomic(path, out)
}

func WriteJSONAtomic(path string, v any) error {
	out, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return writeBytesAtomic(path, out)
}

func writeBytesAtomic(path string, out []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	tmp := path + atomicWriteTmpSuffix
	if err := os.WriteFile(tmp, out, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
