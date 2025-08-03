package event

import "fmt"

type Event struct {
	ID    uint64         `json:"id,omitempty"`
	Ts    uint64         `json:"ts,omitempty"`
	Op    string         `json:"op,omitempty"`
	Table string         `json:"table,omitempty"`
	Row   map[string]any `json:"row,omitempty"`
	Sent  bool           `json:"sent,omitempty"`
}

func (e Event) String() string {
	return fmt.Sprintf(
		"Event[ID=%d, Ts=%d, Op=%s, Table=%s, Sent=%v]",
		e.ID,
		e.Ts,
		e.Op,
		e.Table,
		e.Sent,
	)
}

type Channel struct {
	From   string `yaml:"from"`
	To     string `yaml:"to"`
	Mapper string `yaml:"mapper"`

	Key string `yaml:"-"`
}

func (c Channel) String() string {
	return fmt.Sprintf(
		"Channel[From=%s, To=%s, Mapper=%s, Key=%s]",
		c.From,
		c.To,
		c.Mapper,
		c.Key,
	)
}
