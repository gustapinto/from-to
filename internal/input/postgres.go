package input

import "github.com/gustapinto/from-to/internal"

type PostgresInputConnector struct{}

func (c *PostgresInputConnector) Setup(config any) error {
	// TODO
	return nil
}

func (c *PostgresInputConnector) Listen(out internal.OutputConnector) error {
	for {
		// TODO
	}
}
