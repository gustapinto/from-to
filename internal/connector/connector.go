package connector

type OutputConnector interface {
	Setup(any) error

	Publish(Event) error
}

type InputConnector interface {
	Setup(any) error

	Listen(OutputConnector) error
}
