package data

type DB interface {
	Ping() bool
	Set(KeyValue) error
	Get(string) (string, error)
	Remove(string) error
}
