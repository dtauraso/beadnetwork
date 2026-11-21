package Links

const All = "*"

type Envelope struct {
	To  string
	Msg any
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

type SetS struct{ V int }

type SetTickMs struct{ V int }

type Connect struct {
	From string
	At   Place
}

type Linked struct {
	From string
}

type Put struct {
	From  string
	P     Vec3
	Round int
}

type Moved struct {
	ID         string
	By         Vec3
	At         Vec3
	KeepPicked bool
	Measure    bool
	Jump       bool
	Round      int
}
