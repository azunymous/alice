package data

type DB interface {
	KeyValueDB
	OrderedDB
}

type KeyValueDB interface {
	Ping() bool
	Set(KeyValue) error
	Get(string) (string, error)
	Remove(string) error
}

type OrderedDB interface {
	SetOrdered(KeyValue, int) error
	GetAllOrderedByScore(string) []string
	RemoveOrdered(kv KeyValue) error
}
