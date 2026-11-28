package StepStore

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/dtauraso/beadnetwork/Categories/Node/Links/Store"
)

type Disk struct {
	Step  Step
	Steps map[string]*Step
	Root  string
	Level string
	Nums  []int
	Names map[int]string
	Heads map[int]*Step
}

func (d Disk) Lvl() string { return filepath.Join(d.Root, d.Level) }

func (d Disk) Numbers() []int {
	if len(d.Nums) == 0 {
		return []int{1}
	}
	return d.Nums
}

func (d *Disk) ReadNumbers() {
	d.Nums = ReadStepNumbers(d.Lvl())
	d.Names = map[int]string{}
	for _, n := range d.Numbers() {
		var name string
		if Store.ReadIfExists(filepath.Join(StepDir(d.Lvl(), n), FileStepName), &name) && name != "" {
			d.Names[n] = name
		}
	}
}

func (d Disk) StepLabel(n int) string {
	if name := d.Names[n]; name != "" {
		return strconv.Itoa(n) + " · " + name
	}
	return strconv.Itoa(n)
}

func (d *Disk) NameStep(n int, name string) string {
	if name = strings.TrimSpace(name); name == "step "+strconv.Itoa(n) {
		name = ""
	}
	if d.Root == "" {
		return "steps are kept in the scene's steps/ folder, and this scene has none, so the name was not kept"
	}
	path := filepath.Join(StepDir(d.Lvl(), n), FileStepName)
	var err error
	if name == "" {
		err = os.Remove(path)
		if os.IsNotExist(err) {
			err = nil
		}
	} else {
		err = Store.WriteAtomic(path, name)
	}
	if err != nil {
		return "step " + strconv.Itoa(n) + ": " + err.Error()
	}
	if d.Names == nil {
		d.Names = map[int]string{}
	}
	d.Names[n] = name
	return ""
}

func (d *Disk) StepHead(step int) *Step {
	if d.Heads == nil {
		d.Heads = map[int]*Step{}
	}
	if d.Heads[step] == nil {
		s := Step{}
		if d.Root != "" {
			s = ReadStep(StepDir(d.Lvl(), step))
		}
		d.Heads[step] = &s
	}
	return d.Heads[step]
}

func (d Disk) HeadOf(step int) Step {
	if s := d.Heads[step]; s != nil {
		return *s
	}
	return Step{}
}

func (d Disk) WriteHeads() error {
	for step, s := range d.Heads {
		if err := WriteStep(StepDir(d.Lvl(), step), *s); err != nil {
			return err
		}
	}
	return nil
}

func (d Disk) swapSteps(ns []int, n, by int) (int, string, error) {
	i := slices.Index(ns, n) + by
	if i < 0 || i >= len(ns) {
		return n, fmt.Sprintf("step %d is already the %s step, so it stays", n, map[int]string{-1: "first", 1: "last"}[by]), nil
	}
	k := ns[i]
	a, b := StepDir(d.Lvl(), n), StepDir(d.Lvl(), k)
	swap := a + ".swap"
	if err := os.Rename(a, swap); err != nil {
		return n, "", err
	}
	if err := os.Rename(b, a); err != nil {
		return n, "", err
	}
	return k, "", os.Rename(swap, b)
}

func (d Disk) ShiftSteps(ns []int, after, by int) error {
	var later []int
	for _, k := range ns {
		if k > after {
			later = append(later, k)
		}
	}
	if by > 0 {
		slices.Reverse(later)
	}
	for _, k := range later {
		if err := os.Rename(StepDir(d.Lvl(), k), StepDir(d.Lvl(), k+by)); err != nil {
			return err
		}
	}
	return nil
}

func (d Disk) StepOf(id string) Step {
	if id == "" {
		return d.Step
	}
	if s := d.Steps[id]; s != nil {
		return *s
	}
	return Step{}
}

func (d *Disk) NodeStep(id string) *Step {
	if id == "" {
		return &d.Step
	}
	if d.Steps == nil {
		d.Steps = map[string]*Step{}
	}
	if d.Steps[id] == nil {
		s := Step{}
		if d.Root != "" {
			s = ReadStep(NodePressDir(d.Root, id))
		}
		d.Steps[id] = &s
	}
	return d.Steps[id]
}

func (d Disk) WriteSteps() error {
	if err := WriteStep(PressDir(d.Root), d.Step); err != nil {
		return err
	}
	for id, s := range d.Steps {
		if err := WriteStep(NodePressDir(d.Root, id), *s); err != nil {
			return err
		}
	}
	return d.WriteHeads()
}

func (d Disk) MoveDir(ns []int, from, to int) (string, error) {
	by := 1
	if to < from {
		by = -1
	}
	for at := from; at != to; {
		next, refused, err := d.swapSteps(ns, at, by)
		if refused != "" || err != nil {
			return refused, err
		}
		at = next
	}
	return "", nil
}
