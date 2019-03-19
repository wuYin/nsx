package velar

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

type AddService interface {
	Add(n, diff int) int
}

type AddServiceProxy struct{}

func (s AddServiceProxy) Add(n, diff int) int {
	return n + diff
}

func TestCall(t *testing.T) {
	addService := Service{
		Uri:       "add-service",
		Proxy:     AddServiceProxy{},
		Interface: reflect.TypeOf((*AddService)(nil)).Elem(),
	}

	manager := NewServiceManager([]Service{addService})
	p := ReqPacket{
		ServiceUri: "add-service",
		Method:     "Add",
		Args:       []interface{}{1, 1},
		Timeout:    2 * time.Second,
	}
	resp := manager.Call(p)
	if resp.Ec != 0 {
		t.Fatalf("call failed, ec: %d, em: %s", resp.Ec, resp.Em)
	}

	fmt.Println(p.Method, p.Args, resp.Res) // Add [1 1] 2  // bingo 调用成功
}
