package Stdin

type StdinMsg struct {
	Type string
	Op   string
	Kind string
	Attr string
	Flag string

	Num int

	X, Y float64
}
