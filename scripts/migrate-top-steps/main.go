package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/dtauraso/beadnetwork/Categories/Scene/Scenes"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: go run ./scripts/migrate-top-steps <scene dir>...")
		os.Exit(2)
	}
	for _, scene := range os.Args[1:] {
		if err := migrate(scene); err != nil {
			fmt.Fprintf(os.Stderr, "migrate-top-steps: %s: %v\n", scene, err)
			os.Exit(1)
		}
	}
}

func migrate(scene string) error {
	old := filepath.Join(scene, Scenes.DirOldSeqs)
	if _, err := os.Stat(old); err != nil {
		fmt.Printf("migrate-top-steps: %s: no %s/, nothing to move\n", scene, Scenes.DirOldSeqs)
		return nil
	}
	dst := Scenes.TopStepsDir(scene)
	if _, err := os.Stat(dst); err == nil {
		return fmt.Errorf("both %s/ and %s/ exist, so nothing was moved", Scenes.DirOldSeqs, Scenes.DirTopSteps)
	}
	backup := filepath.Join(".beadnetwork-cache", "backups", filepath.Base(filepath.Clean(scene))+".before-top-steps")
	if _, err := os.Stat(backup); err == nil {
		return fmt.Errorf("%s already exists, so no second backup was taken and nothing was moved", backup)
	}
	if err := os.MkdirAll(filepath.Dir(backup), 0o755); err != nil {
		return err
	}
	if out, err := exec.Command("cp", "-R", scene, backup).CombinedOutput(); err != nil {
		return fmt.Errorf("backup to %s: %v: %s", backup, err, out)
	}
	var open int
	oldOpen := Scenes.ViewFilePath(scene, "sequence.bin")
	Scenes.ReadIfExists(oldOpen, &open)
	if err := os.Rename(old, dst); err != nil {
		return err
	}
	ids, _ := filepath.Glob(filepath.Join(dst, "*", "id.bin"))
	for _, id := range ids {
		if err := os.Remove(id); err != nil {
			return err
		}
	}
	tops := Scenes.TopSteps(scene)
	if len(tops) > 0 {
		if open < 1 {
			open = tops[0]
		}
		if err := Scenes.WriteOpenStep(scene, []int{open}); err != nil {
			return err
		}
	}
	if err := os.Remove(oldOpen); err != nil && !os.IsNotExist(err) {
		return err
	}
	fmt.Printf("migrate-top-steps: %s: top steps %v, open %s, backup %s\n", scene, tops, strconv.Itoa(open), backup)
	return nil
}
