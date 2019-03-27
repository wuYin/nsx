package service

import (
	"encoding/json"
	"fmt"
	"nsx/codec"
	"reflect"
	"testing"
	"time"
)

type AddServiceInterface interface {
	Add(n, diff int) int
	AddAll(base int, diffs []int) int
}

type AddService struct{}

func (s AddService) Add(n, diff int) int {
	return n + diff
}

func (s AddService) AddAll(base int, diffs []int) int {
	for _, d := range diffs {
		base += d
	}
	return base
}

func TestCallAdd(t *testing.T) {
	addService := Service{
		Uri:       "add-service",
		Instance:  AddService{},
		Interface: reflect.TypeOf((*AddServiceInterface)(nil)).Elem(),
	}

	buf, _ := json.Marshal(1)
	manager := NewServiceManager([]Service{addService})
	p := codec.CallReq{
		ServiceUri: "add-service",
		Method:     "Add",
		Args:       []json.RawMessage{buf, buf},
		Timeout:    2 * time.Second,
	}
	resp := manager.Call(p)
	if resp.Ec != 0 {
		t.Fatalf("call failed, ec: %d, em: %s", resp.Ec, resp.Em)
	}

	fmt.Printf("%s %q %q", p.Method, p.Args, resp.Res) // Add [1 1] 2  // bingo 调用成功
}

func TestCallAddAll(t *testing.T) {
	addService := Service{
		Uri:       "add-service",
		Instance:  &AddService{},
		Interface: reflect.TypeOf((*AddServiceInterface)(nil)).Elem(),
	}
	manager := NewServiceManager([]Service{addService})
	buf1, _ := json.Marshal(1)
	buf2, _ := json.Marshal([]int{10, 20})

	p := codec.CallReq{
		ServiceUri: "add-service",
		Method:     "AddAll",
		Args:       []json.RawMessage{buf1, buf2},
		Timeout:    2 * time.Second,
	}

	resp := manager.Call(p)
	fmt.Printf("%s %q %q", p.Method, p.Args, resp.Res)
}
