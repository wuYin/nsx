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

	NewNsxServer("localhost:8080", []string{"localhost:8080"}, services, registry.REG_DEFAULT) // ok
	// NewNsxServer("localhost:8080", []string{"127.0.0.1:2181"}, services, registry.REG_ZK)   // init ok
	fmt.Println(rs.service2Addr) // map[admin-service:localhost:8080]
}

// 服务注册中心
type AdminRegistry struct {
	service2Addr map[string]string
}

func (s AdminRegistry) Register(uri, addr string) error {
	s.service2Addr[uri] = "localhost:8080"
	return nil
}

func (s AdminRegistry) UnRegister(uri, addr string) error {
	delete(s.service2Addr, uri)
	return nil
}

func (s AdminRegistry) GetService(uri string) ([]string, error) {
	addr, ok := s.service2Addr[uri]
	if !ok {
		return nil, fmt.Errorf("default registry: %s not exists", uri)
	}
	return []string{addr}, nil
}
