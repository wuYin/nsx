package codec

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"tron"
)

var (
	CRLF = []byte{'\r', '\n'}
	DELM = []byte{'*'}
)

type ServerCodec struct{}

func NewServerCodec() *ServerCodec {
	return &ServerCodec{}
}

func (c *ServerCodec) ReadPacket(r *bufio.Reader) ([]byte, error) {
	head, _, err := r.ReadLine()
	if err != nil {
		return nil, err
	}
	if !bytes.HasPrefix(head, DELM) {
		return nil, fmt.Errorf("packet: prefix delm %q, not %q", head[:0], DELM)
	}

	readUnit := func(r *bufio.Reader) ([]byte, error) {
		head, err = ReadFullLine(r)
		if err != nil {
			return nil, fmt.Errorf("packet: read fail: %v", err)
		}
		if len(head) < 2 {
			return nil, fmt.Errorf("packet: invalid $ head: %q", head)
		}
		dataLen, err := strconv.Atoi(string(head[1:]))
		if err != nil {
			return nil, fmt.Errorf("packet: atoi failed: %v", err)
		}
		data, err := ReadFullLine(r)
		if err != nil {
			return nil, fmt.Errorf("packet: read fail: %v", err)
		}
		// 校验一下
		if dataLen != len(data) {
			return nil, fmt.Errorf("packet: unmatch data length, head %d, data %d", dataLen, len(data))
		}
		return data, nil
	}
	if _, err := readUnit(r); err != nil {
		return nil, err
	}

	packData, err := readUnit(r) // 读取数据部分
	if err != nil {
		return nil, err
	}

	return packData, nil
}

// 将 Read 读取到的数据直接封装成 packet
func (c *ServerCodec) UnmarshalPacket(reqBuf []byte) (*tron.Packet, error) {
	var cmdReq CmdReq
	if err := json.Unmarshal(reqBuf, &cmdReq); err != nil {
		panic(err)
	}
	return tron.NewRespPacket(cmdReq.Seq, reqBuf), nil
}

// 序列化
// 封装通信协议底层的数据部分
func (c *ServerCodec) MarshalPacket(p tron.Packet) []byte {
	n := len(p.Data)
	buf := bytes.NewBuffer(make([]byte, 0, packBufLen(n)))
	buf.WriteByte('$')
	buf.WriteString(fmt.Sprintf("%d", n))
	buf.Write(CRLF)
	buf.Write(p.Data)
	buf.Write(CRLF)
	return buf.Bytes()
}

// $N CR LF [DATA] CR LF
func packBufLen(dataLen int) int {
	return 1 + len(fmt.Sprintf("%d", dataLen)) + len(CRLF) + dataLen + len(CRLF)
}

// 读取完整的一行
func ReadFullLine(r *bufio.Reader) ([]byte, error) {
	var full []byte
	for {
		l, isPrefix, err := r.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		full = append(full, l...)
		if !isPrefix {
			break
		}
	}
	return full, nil
}
