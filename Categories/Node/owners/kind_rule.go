package owners

import (
	"github.com/dtauraso/wirefold/Categories/Node/nodedrag"
)

type KindRule struct {
	trim nodedrag.Trim

	request nodedrag.Request
}

func (k *KindRule) SetKindRule(trim nodedrag.Trim, request nodedrag.Request) {
	k.trim, k.request = trim, request
}

func (k *KindRule) HasKindRule() bool { return k.trim != nil || k.request != nil }

func (k *KindRule) Trim() nodedrag.Trim { return k.trim }

func (k *KindRule) Request() nodedrag.Request { return k.request }
