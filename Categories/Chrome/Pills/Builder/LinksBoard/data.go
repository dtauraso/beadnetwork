package LinksBoard

import (
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Board"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Sheet"
)

type Act = Sheet.Act

type Lock struct {
	Axis  string
	Value string
	Label string
	On    bool
	Cycle Act
	Press Act
}

type Head struct {
	Play   Lock
	Locks  []Lock
	Picked string
}

type CardRow struct {
	Axis    string
	Start   Board.Stepper
	Dial    *Board.Dial
	Ruler   *Board.Ruler
	A, B, C Board.Stepper
	Total   string
	Why     string
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
	Label    string
	Picked   bool
	Pick     Act
	PickOpen Act
	Gap      Act
	Open     bool
	Fold     Act
	Dial     Board.Dial
	Sum      string
	Hot      bool
	Card     *Card
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
	Drop     Board.Chip
	Menu     Act
	MenuOpen bool
	Head     *Head
	Nodes    []Node
}

type Axis struct {
	Label     string
	Sub       string
	Short     string
	Dial      Board.Dial
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
	Crumbs     []Board.Chip
	CrumbName  Act
	TreeScroll Act
	TreeSY     float32
	InspScroll Act
	InspSY     float32
	TreeSplit  Act
	TreeW      float32
	PathMenu   Act
	PathOpen   bool
	S          Board.Field
	Doors      []Board.Chip
	More       Act
	MoreOpen   bool
	Cols       []Column
	Add        Act
	Insp       Inspector
	InspOpen   bool
	InspTab    Act
	Status     string
	Toast      string
	At         Act
	Undo       Act
	Dismiss    Act
	Pop        *Pop
	Empty      *Empty
	Net        *Net
	New        *NewNode

	Tree      []TreeRow
	TreeFind  *Board.Field
	TreeEmpty string
}
