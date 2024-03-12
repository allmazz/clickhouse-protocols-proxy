package native

func (n *Native) mapResponse(meta []MetaElement, data []interface{}) Response {
	return Response{
		Meta:                   meta,
		Data:                   data,
		Rows:                   len(data),
		RowsBeforeLimitAtLeast: 0,
		Statistics: Statistics{
			Elapsed:   0,
			RowsRead:  0,
			BytesRead: 0,
		},
	}
}
