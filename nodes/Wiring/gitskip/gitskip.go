package gitskip

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

var marked sync.Map

func Mark(path string) {
	abs, err := filepath.Abs(path)
	if err != nil {
		panic(fmt.Sprintf("gitskip.Mark: could not resolve an absolute path for %q: %v", path, err))
	}

	if _, already := marked.Load(abs); already {
		return
	}

	dir := filepath.Dir(abs)

	inTree, err := exec.Command("git", "-C", dir, "rev-parse", "--is-inside-work-tree").Output()
	if err != nil || strings.TrimSpace(string(inTree)) != "true" {
		return
	}

	out, err := exec.Command("git", "-C", dir, "update-index", "--skip-worktree", abs).CombinedOutput()
	if err != nil {
		panic(fmt.Sprintf(
			"gitskip.Mark: git update-index --skip-worktree failed for %s: %v: %s",
			abs, err, strings.TrimSpace(string(out))))
	}

	marked.Store(abs, struct{}{})
}
