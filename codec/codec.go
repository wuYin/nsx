package codec

import (
	"bufio"
	"bytes"
	"fmt"
	"strconv"
	"tron"
)

var (
	CRLF = []byte{'\r', '\n'}
	DELM = []byte{'*'}
)

type VelarCodec struct{}

func (c *VelarCodec) ReadPacket(r *bufio.Reader) ([]byte, error) {
	head, _, err := r.ReadLine()
	if err != nil {
		return nil, err
	}
	if !bytes.HasPrefix(head, DELM) {
		return nil, fmt.Errorf("packet: prefix delm %q, not %q", head[:0], DELM)
	}

	buf := bytes.NewBuffer(nil)
	for {
		line, isPrefix, err := r.ReadLine()
		fmt.Println("line", line)
		if err != nil {
			return nil, fmt.Errorf("packet: read fail: %v", err)
		}
		if _, err = buf.Write(line[1:]); err != nil { // 丢弃 *
			return nil, fmt.Errorf("packet: buf write fail: %v", err)
		}
		if !isPrefix {
			break
		}
	}
	dataLen, err := strconv.Atoi(buf.String())
	if err != nil {
		return nil, fmt.Errorf("packet: %q atoi failed: %v", buf.String(), err)
	}
	newBuf := bytes.NewBuffer(make([]byte, 0, dataLen))
	readLen := 0
	for {
		line, isPrefix, err := r.ReadLine()
		if err != nil {
			return nil, fmt.Errorf("packet: read fail: %v", err)
		}
		n, err := newBuf.Write(line)
		if err != nil {
			return nil, fmt.Errorf("packet: buf write fail: %v", err)
		}
		readLen += n
		if !isPrefix {
			break
		}
	}

	return newBuf.Bytes(), nil
}

// 将 Read 读取到的数据直接封装成 packet
func (c *VelarCodec) UnmarshalPacket(buf []byte) (*tron.Packet, error) {
	return tron.NewPacket(buf), nil
}

// 序列化
// 封装通信协议底层的数据部分
func (c *VelarCodec) MarshalPacket(p tron.Packet) []byte {
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
