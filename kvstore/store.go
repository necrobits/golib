package kvstore

type Data interface{}

type KvStore interface {
	Get(key string) (Data, error)
	Set(key string, value Data) error
	Delete(key string) error
	Transaction(func(KvStore) error) error
}
