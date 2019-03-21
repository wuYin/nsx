package codec

import (
	"bufio"
	"bytes"
	"fmt"
	"testing"
	"tron"
)

func TestPacket(t *testing.T) {
	c := &VelarCodec{}
	p1 := tron.NewPacket([]byte("a"))
	b := c.MarshalPacket(*p1) // "$1\r\na\r\n"

	// 手动模拟协议
	b = append([]byte("*2\r\n"), b...) //  "*2\r\n$1\r\na\r\n"

	r := bufio.NewReader(bytes.NewReader(b))
	p2, err := c.ReadPacket(r)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%q\n", p2)
}
