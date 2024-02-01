package kvstore

import "context"

type Data interface{}

type KvStore interface {
	Get(ctx context.Context, key string) (Data, error)
	Set(ctx context.Context, key string, value Data) error
	Delete(ctx context.Context, key string) error
	Transaction(ctx context.Context, fn func(tx KvStore) error) error
}
