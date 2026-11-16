package Scenes

import (
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
)

const (
	DirSequences = "sequences"
	FileSeqName  = "name.bin"
	FileSeqID    = "id.bin"
)

func SequenceID(scenePath string, k int) int {
	var id int
	ReadIfExists(filepath.Join(SequenceDir(scenePath, k), FileSeqID), &id)
	return id
}

func nextSequenceID(scenePath string) int {
	most := 0
	for _, k := range Sequences(scenePath) {
		most = max(most, SequenceID(scenePath, k))
	}
	return most + 1
}

func EnsureSequenceIDs(scenePath string) error {
	seqs := Sequences(scenePath)
	used := map[int]bool{}
	for _, k := range seqs {
		used[SequenceID(scenePath, k)] = true
	}
	for _, k := range seqs {
		if SequenceID(scenePath, k) > 0 {
			continue
		}
		id := k
		if used[id] {
			id = nextSequenceID(scenePath)
		}
		if err := WriteAtomic(filepath.Join(SequenceDir(scenePath, k), FileSeqID), id); err != nil {
			return err
		}
		used[id] = true
	}
	return nil
}

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
	id := SequenceID(scenePath, k)
	if id == 0 {
		id = k
	}
	name := "sequence " + strconv.Itoa(id)
	ReadIfExists(filepath.Join(SequenceDir(scenePath, k), FileSeqName), &name)
	return name
}

func WriteOpenSequence(scenePath string, k int) error {
	return WriteAtomic(SequenceFilePath(scenePath), k)
}

func WriteSequenceName(scenePath string, k int, name string) error {
	return WriteAtomic(filepath.Join(SequenceDir(scenePath, k), FileSeqName), name)
}

func CopySequence(scenePath string, from int) (int, error) {
	seqs, id := Sequences(scenePath), nextSequenceID(scenePath)
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
	if err := WriteAtomic(filepath.Join(dst, FileSeqID), id); err != nil {
		return 0, err
	}
	return to, WriteSequenceName(scenePath, to, "sequence "+strconv.Itoa(id))
}

func RemoveSequence(scenePath string, k int) error {
	if err := os.RemoveAll(SequenceDir(scenePath, k)); err != nil {
		return err
	}
	for i, s := range Sequences(scenePath) {
		if want := i + 1; s != want {
			if err := os.Rename(SequenceDir(scenePath, s), SequenceDir(scenePath, want)); err != nil {
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
	return to, nil
}
