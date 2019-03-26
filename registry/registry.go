package registry

import (
	"nsx/service"
)

type Registry interface {
	RegisterService(uri string) error
	UnRegisterService(uri string) error
	GetService(uri string) string
}

type RegistryCenter struct {
	addr     string
	register Registry
	services []service.Service
}

// 将所有 service 注册到 register 上
func NewRegistryCenter(centerAddr string, services []service.Service) *RegistryCenter {
	var reg Registry = NewDefaultRegistry(centerAddr)
	return &RegistryCenter{
		addr:     centerAddr,
		register: reg,
		services: services,
	}
}

func (c *RegistryCenter) RegisterAllService() {
	for _, s := range c.services {
		c.register.RegisterService(s.Uri)
	}
}
