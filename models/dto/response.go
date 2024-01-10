package dto

import "encoding/json"

type Response struct {
	Result string          `json:"result"`
	Data   json.RawMessage `json:"data"`
	Error  string          `json:"error"`
}

func (r *Response) Update(result string, data json.RawMessage, error string) {
	r.Result = result
	r.Data = data
	r.Error = error
}
