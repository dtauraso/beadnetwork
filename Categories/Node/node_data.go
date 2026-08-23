package Node

type NodeData struct {
	Label  string
	Init   []int
	Repeat bool
	State  map[string]int

	SendRules map[string]string
}
