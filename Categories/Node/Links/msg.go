package Links

const All = "*"

type Envelope struct {
	To  string
	Msg any
}

type SetStart struct {
	To string
	K  string
	V  float64
}

type Press struct {
	To   string
	K    string
	Sign int
}

type SetIncrement struct {
	To string
	K  string
	V  float64
}

type SetLocking struct {
	To string
	On bool
}

type StepAll struct {
	Round   int
	H       int64
	K       string
	Sign    int
	LockAll Lock
}

const KBoth = "both"

type PickReq struct{ From string }

type PickSet struct {
	From string
	On   bool
}

type Unpick struct{ To string }

type PickDropped struct{ From string }

type SetIdx struct {
	K string
	V float64
}

type SetS struct{ V int }

type SetTickMs struct{ V int }

type Connect struct {
	To string
	At Place
}

type Drop struct{ To string }

type Linked struct{ From string }

type Unlinked struct{ From string }

type Put struct {
	From  string
	P, D  Vec3
	Keep  string
	Round int
	Step  bool
	Final bool
}

type Moved struct {
	ID         string
	At         Place
	Keep       string
	KeepPicked bool
}
