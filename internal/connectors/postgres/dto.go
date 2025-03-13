package postgres

type SetupParamsTable struct {
	Name      string
	KeyColumn string
	Topic     string
}

type SetupParams struct {
	DSN         string
	PollSeconds int64
	Tables      []SetupParamsTable
}
