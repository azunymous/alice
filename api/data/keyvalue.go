package data

import "fmt"

type KeyValue interface {
	Key() string
	fmt.Stringer
}

type KV struct {
	k string
	v string
}

func NewKeyValuePair(key, value string) KeyValue {
	return KV{key, value}
}

func (p KV) Key() string {
	return p.k
}

func (p KV) String() string {
	return p.v
}
