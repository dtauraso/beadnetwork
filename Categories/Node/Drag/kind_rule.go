package Drag

import ()

type KindRule struct {
	trim Trim

	request Request
}

func (k *KindRule) SetKindRule(trim Trim, request Request) {
	k.trim, k.request = trim, request
}


func (k *KindRule) Trim() Trim { return k.trim }

func (k *KindRule) Request() Request { return k.request }
