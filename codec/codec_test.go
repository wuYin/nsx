package codec

import (
	"bufio"
	"encoding/json"
	"fmt"
	"gopkg.in/redis.v3"
	"log"
	"net"
	"strings"
	"testing"
	"tron"
)

func TestPacket(t *testing.T) {
	addr, _ := net.ResolveTCPAddr("tcp4", "localhost:2333")
	l, err := net.ListenTCP("tcp4", addr)
	if err != nil {
		log.Fatal(err)
	}

	codec := &ServerCodec{}
	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				fmt.Println(err)
				continue
			}
			buf := make([]byte, 1024)
			n, err := conn.Read(buf)
			if err != nil {
				fmt.Println(err)
				continue
			}
			// cmd := string(buf[:n])
			// fmt.Printf("[cmd]: `%s`\n", cmd)

			pack, err := codec.UnmarshalPacket(buf[:n])
			if err != nil {
				fmt.Println(err)
			}

			r := bufio.NewReader(strings.NewReader(string(pack.Data)))
			b, err := codec.ReadPacket(r)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Printf("[read packet]: %s\n", string(b))

			respPack := tron.NewRespPacket(0, []byte("A"))
			b = codec.MarshalPacket(*respPack)

			fmt.Printf("[resp packet]: `%s`\n", b)

			n, err = conn.Write(b)
			fmt.Println(n, err)
		}
	}()

	go func() {
		cli := redis.NewClient(&redis.Options{
			Addr:     "localhost:2333",
			DB:       0,
			Password: "",
		})

		cmd := CmdReq{
			ServiceUri: "add-service",
			Method:     "Add",
			Args:       []interface{}{1, 2},
		}
		b, _ := json.Marshal(cmd)
		res, err := cli.Get(string(b)).Result()
		if err != nil {
			fmt.Println("get failed: ", err)
			return
		}
		fmt.Println("res", res)
	}()

	select {}
}

func TestReadFullLine(t *testing.T) {
	s := "a\r\nb\nc"
	r := bufio.NewReader(strings.NewReader(s))
	buf, err := ReadFullLine(r)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%q\n", buf)
	buf, err = ReadFullLine(r)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%q\n", buf)
}
