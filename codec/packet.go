package codec

import (
	"encoding/json"
	"fmt"
	"time"
	"tron"
)

type CallReq struct {
	ServiceUri string            `json:"uri"`
	Method     string            `json:"method"`
	Args       []json.RawMessage `json:"args"`
	Timeout    time.Duration     `json:"-"`
}

type CallResp struct {
	Seq int32           `json:"seq"`
	Ec  int             `json:"ec"`
	Em  string          `json:"em"`
	Res json.RawMessage `json:"res"`
}

type CmdReq struct {
	Seq        int32         `json:"seq"`
	ServiceUri string        `json:"uri"`
	Method     string        `json:"method"`
	Args       []interface{} `json:"args"`
}

func CmdReq2CallReq(rawData []byte) (*CallReq, error) {
	var cmd CmdReq
	if err := json.Unmarshal(rawData, &cmd); err != nil {
		return nil, err
	}

	var rawArgs []json.RawMessage
	for _, arg := range cmd.Args {
		buf, err := json.Marshal(arg)
		if err != nil {
			panic(fmt.Sprintf("cmd to call req: marshal %+v failed: %v", arg, err))
		}
		rawArgs = append(rawArgs, buf)
	}

	call := &CallReq{
		ServiceUri: cmd.ServiceUri,
		Method:     cmd.Method,
		Args:       rawArgs,
	}
	return call, nil
}

func CallResp2Packet(callResp CallResp, reqPack *tron.Packet) *tron.Packet {
	data, _ := json.Marshal(callResp)
	respPack := tron.NewRespPacket(reqPack.Seq(), data)
	return respPack
}
