package types

type CircuitBreaker interface {
	GetState() string
	CanExecute() bool
	Execute(err error)
}
