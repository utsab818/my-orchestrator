package store

// we can use `any` or `interface{}` to define any type we want.
type Store interface {
	Put(key string, value any) error
	Get(key string) (any, error)
	List() (any, error)
	Count() (int, error)
}
