package codec

import "time"

type CallReq struct {
	ServiceUri string        `json:"uri"`
	Method     string        `json:"method"`
	Args       []interface{} `json:"args"`
	Timeout    time.Duration `json:"-"`
}

type CallResp struct {
	Ec  int         `json:"ec"`
	Em  string      `json:"em"`
	Res interface{} `json:"res"`
}
