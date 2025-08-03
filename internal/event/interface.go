package event

type Mapper interface {
	Map(Event) ([]byte, error)
}

type Listener interface {
	Listen(func(event Event, channels []Channel) error) error
}

type Publisher interface {
	Publish(Event, []byte, string) error
}
