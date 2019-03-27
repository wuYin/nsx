package registry

import (
	"nsx/service"
)

type Registry interface {
	Register(uri, addr string) error         // uri 新增实例
	UnRegister(uri, addr string) error       // 将 addr 实例从 uri 实例列表下线
	GetService(uri string) ([]string, error) // 获取指定名字的服务
}

type RegistryCenter struct {
	addr     string // 业务服务器地址
	registry Registry
	services []service.Service
}

const (
	REG_DEFAULT = 0 // 默认字典注册中心
	REG_ZK      = 1 // zk 注册中心
)

// 将所有 service 注册到目标的 map 上，仅单机使用
func NewDefaultRegistryCenter(serverAddr, regAddr string, services []service.Service) *RegistryCenter {
	var reg Registry = NewDefaultRegistry(regAddr)
	return &RegistryCenter{
		addr:     serverAddr,
		registry: reg,
		services: services,
	}
}

// 集群使用，但暂时还是
func NewZKRegistryCenter(serverAddr string, zkServers []string, services []service.Service) *RegistryCenter {
	r := NewZKRegistry(zkServers)
	return &RegistryCenter{
		addr:     serverAddr, // 业务服务器地址
		registry: r,
		services: services,
	}
}

func (c *RegistryCenter) RegisterAllService() {
	for _, s := range c.services {
		c.registry.Register(s.Uri, c.addr)
	}
}
