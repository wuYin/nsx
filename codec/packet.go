package codec

import (
	"encoding/json"
	"time"
	"tron"
)

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

type CmdReq struct {
	ServiceUri string        `json:"uri"`
	Method     string        `json:"method"`
	Args       []interface{} `json:"args"`
}

func CmdReq2CallReq(rawData []byte) (*CallReq, error) {
	var cmd CmdReq
	if err := json.Unmarshal(rawData, &cmd); err != nil {
		return nil, err
	}

	call := &CallReq{
		ServiceUri: cmd.ServiceUri,
		Method:     cmd.Method,
		Args:       cmd.Args,
	}
	return call, nil
}

func CallResp2Packet(resp CallResp, reqPack *tron.Packet) *tron.Packet {
	data, _ := json.Marshal(resp)
	respPack := tron.NewPacket(data)
	respPack.Header.Seq = reqPack.Header.Seq
	return respPack
}
