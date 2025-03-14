package postgres

import "github.com/gustapinto/from-to/internal/event"

type SetupParamsTable struct {
	Name          string
	KeyColumn     string
	EventMetadata event.Config
}

type SetupParams struct {
	DSN         string
	PollSeconds int64
	Tables      []SetupParamsTable
}
