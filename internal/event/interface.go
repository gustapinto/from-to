package event

type Mapper interface {
	Map(Event) ([]byte, error)
}

type Listener interface {
	Listen(callback func(event Event) error) error
}

type Publisher interface {
	Publish(meta Event, payload []byte) error
}
