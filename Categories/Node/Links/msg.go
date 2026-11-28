package Links

const All = "*"

type Envelope struct {
	To  string
	Msg any
}

type StepAll struct {
	Round int
	H     int64
}

const KBoth = "both"

type Amount struct {
	From string
	D    Ints
	Play bool
}

type PlayPart struct {
	From      string
	Press, Of int
}

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
	Round      int
}
