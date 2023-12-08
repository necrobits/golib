package kvstore

type Data interface{}

type KvStore interface {
	Get(key string) (Data, error)
	Set(key string, value Data) error
	SetMany(data map[string]Data) error
	Delete(key string) error
	DeleteMany(keys []string) error
}
