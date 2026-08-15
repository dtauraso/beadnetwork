package gitskip

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

var marked sync.Map

var reported sync.Map

func Mark(path string) {
	abs, err := filepath.Abs(path)
	if err != nil {
		reportOnce(path, fmt.Sprintf("could not resolve an absolute path: %v", err))
		return
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
		reportOnce(abs, fmt.Sprintf("%v: %s", err, strings.TrimSpace(string(out))))
		return
	}

	marked.Store(abs, struct{}{})
}

func reportOnce(path, detail string) {
	if _, seen := reported.LoadOrStore(path, struct{}{}); seen {
		return
	}
	fmt.Fprintf(os.Stderr,
		"gitskip.Mark: git update-index --skip-worktree failed for %s (%s); the file will read as MODIFIED in git until a later write succeeds\n",
		path, detail)
}
