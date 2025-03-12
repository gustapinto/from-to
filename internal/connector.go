package internal

type Row struct {
	Key   string
	Topic string
	Data  map[string]any
}

type OutputConnector interface {
	Setup(any) error

	Publish(Row) error
}

type InputConnector interface {
	Setup(any) error

	Listen(OutputConnector) error
}
