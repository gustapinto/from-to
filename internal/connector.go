package internal

type EventMetadata struct {
	Key      string
	KeyValue string
	Topic    string
}

type Event struct {
	ID       int64          `json:"id,omitempty"`
	Ts       int64          `json:"ts,omitempty"`
	Op       string         `json:"op,omitempty"`
	Table    string         `json:"table,omitempty"`
	Row      map[string]any `json:"row,omitempty"`
	Metadata EventMetadata  `json:"-"`
}

type OutputConnector interface {
	Setup(any) error

	Publish(Event) error
}

type InputConnector interface {
	Setup(any) error

	Listen(OutputConnector) error
}
