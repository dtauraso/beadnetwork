package streamwire

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

const StreamKindView = "view"

const StreamKindEdge = "edge"

const StreamKindNode = "node"

const StreamKindInterior = "interior"

const StreamKindBead = "bead"

type StreamFDs map[string]int

func ParseStreamFDs(env string) StreamFDs {
	out := StreamFDs{}
	if env == "" {
		return out
	}
	for _, part := range strings.Split(env, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		kv := strings.SplitN(part, ":", 2)
		if len(kv) != 2 {
			continue
		}
		n, err := strconv.Atoi(strings.TrimSpace(kv[1]))
		if err != nil {
			continue
		}
		out[strings.TrimSpace(kv[0])] = n
	}
	return out
}

func (m StreamFDs) FD(kind string, row int) (int, bool) {
	base, ok := m[kind]
	if !ok {
		return 0, false
	}
	return base + row, true
}

func (m StreamFDs) Open(kind string, row int) (*os.File, bool) {
	fdNum, ok := m.FD(kind, row)
	if !ok {
		return nil, false
	}
	return os.NewFile(uintptr(fdNum), fmt.Sprintf("%s-fd%d", kind, fdNum)), true
}
