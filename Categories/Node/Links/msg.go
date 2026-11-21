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
	KeepPicked bool
	Measure    bool
	Jump       bool
	Round      int
}

type PlayFrame struct {
	Run  int
	K, N int
	Stop bool
}

type PlayAt struct {
	Run  int
	From string
	P    Vec3
	K, N int
}

func (m PlayFrame) Frac() float64 { return float64(m.K) / float64(max(1, m.N)) }

func (m PlayAt) Frac() float64 { return float64(m.K) / float64(max(1, m.N)) }
