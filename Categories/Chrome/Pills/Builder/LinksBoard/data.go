package LinksBoard

import (
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Board"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Sheet"
)

type Act = Sheet.Act

type Crumb struct {
	Label  string
	Short  string
	Tag    string
	Act    Act
	Edit   Act
	Rename Act
}

type Lock struct {
	Label string
	On    bool
	Cycle Act
	Press Act
}

type Head struct {
	Locks  []Lock
	Picked string
}

type CardRow struct {
	Axis    string
	Start   Board.Stepper
	Dial    *Board.Dial
	DialAct Act
	Ruler   *Board.Ruler
	A, B    Board.Stepper
	Total   string
}

type Card struct {
	Heads     [3]string
	StartHead string
	Rows      []CardRow
	Note      string
	Fixed     string
	Foot      string
}

type Link struct {
	Label  string
	Picked bool
	Pick   Act
	Open   bool
	Fold   Act
	Dial   Board.Dial
	Sum    string
	Tip    string
	Hot    bool
	Card   *Card
}

type Node struct {
	Label, Sub string
	Open       bool
	Fold       Act
	Links      []Link
}

type Column struct {
	Label    string
	Short    string
	Sub      string
	Open     bool
	Shown    bool
	Fold     Act
	Show     Act
	Progress string
	Name     Board.Field
	Tools    []Board.Chip
	Head     *Head
	Nodes    []Node
}

type Axis struct {
	Label     string
	Sub       string
	Short     string
	Dial      Board.Dial
	DialAct   Act
	DialSize  float32
	Turn      Board.Stepper
	IncTag    string
	Inc       Board.Stepper
	Lines     []string
	Ruler     *Board.Ruler
	RulerTags [3]string
	Ends      string
	Fixed     string
}

type Inspector struct {
	Title  string
	Sub    string
	Moves  string
	Switch Board.Chip
	Note   string
	Axes   []Axis
	Hint   string
}

type PopField struct {
	Label, What, Was string
	Field            Board.Field
}

type Pop struct {
	Tag      string
	Fields   []PopField
	Apply    Act
	Cancel   Act
	CanApply bool
}

type Empty struct {
	Title, Sub string
	Door       Board.Chip
}

type Data struct {
	Crumbs     []Crumb
	S          Board.Field
	SNote      string
	Doors      []Board.Chip
	DoorsShort []string
	Cols       []Column
	Add        Act
	Insp       Inspector
	InspOpen   bool
	InspTab    Act
	Status     string
	Ghost      string
	Toast      string
	Undo       Act
	Dismiss    Act
	Pop        *Pop
	Empty      *Empty
}
