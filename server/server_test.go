package server

import (
	"fmt"
	"nsx/registry"
	"nsx/service"
	"reflect"
	"testing"
)

func TestServer(t *testing.T) {
	rs := AdminRegistry{
		service2Addr: make(map[string]string),
	}
	services := []service.Service{
		{
			Uri:       service.SERVICE_ADMIN,
			Instance:  rs,
			Interface: reflect.TypeOf((*registry.Registry)(nil)).Elem(),
		},
	}

	NewNsxServer("localhost:8080", services)
	fmt.Println(rs.service2Addr) // map[registry-service:localhost:8080]
}

// 服务注册中心
type AdminRegistry struct {
	service2Addr map[string]string
}

func (s AdminRegistry) RegisterService(uri string) error {
	s.service2Addr[uri] = "localhost:8080"
	return nil
}

func (s AdminRegistry) UnRegisterService(uri string) error {
	delete(s.service2Addr, uri)
	return nil
}

func (s AdminRegistry) GetService(uri string) (addr string) {
	return s.service2Addr[uri]
}
