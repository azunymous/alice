package data

type DB interface {
	Ping() bool
	Add(KeyValue) error
	Get(string) (string, error)
	Remove(string) error
}
