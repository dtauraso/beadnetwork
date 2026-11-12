package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Pills/Builder"
	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: go run ./scripts/migrate-steps <scene dir>...")
		os.Exit(2)
	}
	for _, scene := range os.Args[1:] {
		seqs, _ := filepath.Glob(filepath.Join(scene, "sequences", "*"))
		for _, root := range append([]string{scene}, seqs...) {
			if err := migrate(root); err != nil {
				fmt.Fprintf(os.Stderr, "migrate-steps: %s: %v\n", root, err)
				os.Exit(1)
			}
		}
	}
}

func migrate(root string) error {
	old := filepath.Join(root, "step")
	if _, err := os.Stat(old); err != nil {
		return nil
	}
	if _, err := os.Stat(Builder.StepsDir(root)); err == nil {
		return fmt.Errorf("both step/ and %s/ exist, so nothing was moved", Builder.DirSteps)
	}
	if err := moveLocks(old, Builder.PressDir(root)); err != nil {
		return err
	}
	entries, _ := os.ReadDir(old)
	for _, e := range entries {
		nodeDir := filepath.Join(root, Builder.DirNodes, e.Name())
		if _, err := os.Stat(nodeDir); err != nil || !e.IsDir() { // path-resolution-ok: old step/ held one directory per node
			continue
		}
		if err := moveLocks(filepath.Join(old, e.Name()), filepath.Join(nodeDir, Builder.DirPress)); err != nil {
			return err
		}
	}
	nums := map[int]bool{1: true, 2: true}
	for _, sub := range []string{"heads", "picks", "clones"} {
		es, _ := os.ReadDir(filepath.Join(old, sub))
		for _, e := range es {
			if n, err := strconv.Atoi(strings.TrimPrefix(e.Name(), "step-")); err == nil && n > 0 {
				nums[n] = true
			}
		}
	}
	for n := range nums {
		if err := migrateStep(root, old, n); err != nil {
			return err
		}
	}
	fmt.Printf("migrate-steps: %s: steps %v\n", root, Builder.ReadStepNumbers(root))
	return os.RemoveAll(old)
}

func migrateStep(root, old string, n int) error {
	name := "step-" + strconv.Itoa(n)
	dst := filepath.Join(Builder.StepsDir(root), strconv.Itoa(n))
	nodes := filepath.Join(dst, Builder.DirNodes)
	if err := os.MkdirAll(nodes, 0o755); err != nil {
		return err
	}
	if err := moveLocks(filepath.Join(old, "heads", name), dst); err != nil {
		return err
	}
	clones := filepath.Join(old, "clones", name)
	tos, _ := os.ReadDir(clones)
	for _, to := range tos {
		if !to.IsDir() { // path-resolution-ok: old clones held one directory per target node
			continue
		}
		if err := os.Rename(filepath.Join(clones, to.Name()), filepath.Join(nodes, to.Name())); err != nil {
			return err
		}
	}
	base, _ := filepath.Glob(filepath.Join(nodes, "*", Links.DirLinks, "*"))
	drag, _ := filepath.Glob(filepath.Join(nodes, "*", Links.DirDrag, Links.DirLinks, "*"))
	for _, dir := range append(base, drag...) {
		var on bool
		Links.ReadIfExists(filepath.Join(old, "picks", name, filepath.Base(dir)+".bin"), &on)
		if err := Links.WriteAtomic(filepath.Join(dir, Links.FilePicked), on); err != nil {
			return err
		}
	}
	return nil
}

func moveLocks(src, dst string) error {
	files, _ := filepath.Glob(filepath.Join(src, "lock-*.bin"))
	if len(files) == 0 {
		return nil
	}
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return err
	}
	for _, f := range files {
		if err := os.Rename(f, filepath.Join(dst, filepath.Base(f))); err != nil {
			return err
		}
	}
	return nil
}
