package Links

const All = "*"

type Envelope struct {
	To  string
	Msg any
}

type SetLink struct {
	From      string
	Start     Start
	Delta     Ticks
	Increment Ticks
}

type StepAll struct {
	Round   int
	H       int64
	K       string
	Sign    int
	From    string
	LockAll Lock
	FixedR  []string
}

const KBoth = "both"

type PickReq struct{ From string }

type PickSet struct {
	From string
	On   bool
}

type Unpick struct{ From string }

type SetIdx struct {
	K string
	V float64
}

type SetS struct {
	V    int
	Note string
}

type SSet struct {
	From string
	S    int
}

type SetTickMs struct{ V int }

type Connect struct {
	From string
	At   Place
}

type Linked struct {
	From string
	At   Place
}

type Put struct {
	From  string
	P     Vec3
	Round int
}

type Moved struct {
	ID         string
	At         Place
	KeepPicked bool
	Round      int
}
