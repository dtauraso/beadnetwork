package Scenes

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strconv"
)

const (
	DirSequences = "sequences"
	FileSeqName  = "name.bin"
)

func SequencesDir(scenePath string) string { return filepath.Join(scenePath, DirSequences) }

func SequenceDir(scenePath string, k int) string {
	return filepath.Join(SequencesDir(scenePath), strconv.Itoa(k))
}

func SequenceFilePath(scenePath string) string { return ViewFilePath(scenePath, "sequence.bin") }

func Sequences(scenePath string) []int {
	entries, err := os.ReadDir(SequencesDir(scenePath))
	if err != nil {
		return nil
	}
	var out []int
	for _, e := range entries {
		k, err := strconv.Atoi(e.Name())
		if err == nil && k > 0 && e.IsDir() { // path-resolution-ok: a sequence is a numbered directory under sequences/; other entries are skipped
			out = append(out, k)
		}
	}
	sort.Ints(out)
	return out
}

func OpenSequence(scenePath string) int {
	seqs := Sequences(scenePath)
	if len(seqs) == 0 {
		return 0
	}
	var k int
	ReadIfExists(SequenceFilePath(scenePath), &k)
	if slices.Contains(seqs, k) {
		return k
	}
	return seqs[0]
}

func NodesRoot(scenePath string) string {
	if k := OpenSequence(scenePath); k > 0 {
		return SequenceDir(scenePath, k)
	}
	return scenePath
}

func OpenSFilePath(scenePath string) string { return SFilePath(NodesRoot(scenePath)) }

func SequenceName(scenePath string, k int) string {
	name := "sequence " + strconv.Itoa(k)
	ReadIfExists(filepath.Join(SequenceDir(scenePath, k), FileSeqName), &name)
	return name
}

func WriteOpenSequence(scenePath string, k int) error {
	return WriteAtomic(SequenceFilePath(scenePath), k)
}

func WriteSequenceName(scenePath string, k int, name string) error {
	return WriteAtomic(filepath.Join(SequenceDir(scenePath, k), FileSeqName), name)
}

var numbered = regexp.MustCompile(`^sequence \d+$`)

func CopySequence(scenePath string, from int) (int, error) {
	seqs := Sequences(scenePath)
	to := 1
	if len(seqs) > 0 {
		to = seqs[len(seqs)-1] + 1
	}
	src, dst := SequenceDir(scenePath, from), SequenceDir(scenePath, to)
	err := filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, path)
		target := filepath.Join(dst, rel)
		if d.IsDir() { // path-resolution-ok: copying a sequence directory tree
			return os.MkdirAll(target, 0o755)
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, raw, 0o644)
	})
	if err != nil {
		return 0, err
	}
	return to, WriteSequenceName(scenePath, to, "sequence "+strconv.Itoa(len(seqs)+1))
}

func RemoveSequence(scenePath string, k int) error {
	if err := os.RemoveAll(SequenceDir(scenePath, k)); err != nil {
		return err
	}
	for i, s := range Sequences(scenePath) {
		want := i + 1
		if s != want {
			if err := os.Rename(SequenceDir(scenePath, s), SequenceDir(scenePath, want)); err != nil {
				return err
			}
		}
		if numbered.MatchString(SequenceName(scenePath, want)) {
			if err := WriteSequenceName(scenePath, want, "sequence "+strconv.Itoa(want)); err != nil {
				return err
			}
		}
	}
	return nil
}

func InsertSequence(scenePath string, after int, drop string) (int, error) {
	end, err := CopySequence(scenePath, after)
	if err != nil {
		return 0, err
	}
	if drop != "" {
		if err := os.RemoveAll(filepath.Join(SequenceDir(scenePath, end), drop)); err != nil {
			return 0, err
		}
	}
	to := after + 1
	if end != to {
		tmp := filepath.Join(SequencesDir(scenePath), "inserting")
		if err := os.Rename(SequenceDir(scenePath, end), tmp); err != nil {
			return 0, err
		}
		for s := end - 1; s >= to; s-- {
			if err := os.Rename(SequenceDir(scenePath, s), SequenceDir(scenePath, s+1)); err != nil {
				return 0, err
			}
		}
		if err := os.Rename(tmp, SequenceDir(scenePath, to)); err != nil {
			return 0, err
		}
	}
	for _, s := range Sequences(scenePath) {
		if numbered.MatchString(SequenceName(scenePath, s)) {
			if err := WriteSequenceName(scenePath, s, "sequence "+strconv.Itoa(s)); err != nil {
				return 0, err
			}
		}
	}
	return to, nil
}
