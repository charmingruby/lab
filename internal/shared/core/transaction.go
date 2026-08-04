package core

type TransactionManager[T any] interface {
	Transact(func(tx T) error) error
}
