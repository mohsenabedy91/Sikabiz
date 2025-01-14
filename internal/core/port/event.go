package port

type Event interface {
	Name() string
	Publish(message interface{}) error
	Consume(message []byte) error
	Register()
}
