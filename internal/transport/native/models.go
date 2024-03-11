package native

type Query struct {
	Database string

	Username string
	Password string

	Query  string
	Params map[string]any
}

type Statistics struct {
	Elapsed   float64 `json:"elapsed"`
	RowsRead  int     `json:"rows_read"`
	BytesRead int     `json:"bytes_read"`
}

type MetaElement struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type Response struct {
	Meta                   []MetaElement `json:"meta"`
	Data                   []interface{} `json:"data"`
	Rows                   int           `json:"rows"`
	RowsBeforeLimitAtLeast int           `json:"rows_before_limit_at_least"`
	Statistics             Statistics    `json:"statistics"`
}
