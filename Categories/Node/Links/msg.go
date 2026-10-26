package Links

const All = "*"

type Envelope struct {
	To  string
	Msg any
}

type SetStart struct {
	From string
	K    string
	V    float64
}

type Press struct {
	From string
	K    string
	Sign int
}

type SetIncrement struct {
	From string
	K    string
	V    float64
}

type SetLocking struct {
	From string
	On   bool
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

type Unpick struct{ From string }

type SetIdx struct {
	K string
	V float64
}

type SetS struct{ V int }

type SetTickMs struct{ V int }

type Connect struct {
	From string
	At   Place
}

type Drop struct{ From string }

type Linked struct {
	From string
	At   Place
}

type Unlinked struct{ From string }

type Put struct {
	From  string
	P     Vec3
	Round int
	Final bool
}

type Moved struct {
	ID         string
	At         Place
	KeepPicked bool
	Round      int
}
