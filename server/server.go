package server

import (
	"fmt"
	"log"
	"nsx/codec"
	"nsx/registry"
	"nsx/service"
	"time"
	"tron"
)

type NsxServer struct {
	server  *tron.Server
	manager *service.ServiceManager
}

func NewNsxServer(serverAddr string, regServers []string, services []service.Service, regType int) *NsxServer {
	conf := tron.NewDefaultConf(1 * time.Minute)
	s := &NsxServer{}
	s.manager = service.NewServiceManager(services)
	s.server = tron.NewServer(serverAddr, conf, codec.NewServerCodec(), s.packetHandler)

	if err := s.server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}

	var regCenter *registry.RegistryCenter
	switch regType {
	case registry.REG_DEFAULT:
		regCenter = registry.NewDefaultRegistryCenter(serverAddr, regServers[0], services)
	case registry.REG_ZK:
		regCenter = registry.NewZKRegistryCenter(serverAddr, regServers, services) // 单机测
	default:
		s.server.Shutdown()
		panic(fmt.Sprintf("invalid registry type: %d", regType))
	}
	regCenter.RegisterAllService() // 注册所有服务

	return s
}

// 执行本地调用
func (s NsxServer) packetHandler(worker *tron.Client, p *tron.Packet) {
	callReq, err := codec.CmdReq2CallReq(p.Data)
	if err != nil {
		fmt.Println("packetHandler", err)
	}
	callReq.Timeout = 5 * time.Second

	// 执行调用
	resp := s.manager.Call(*callReq)
	if resp.Ec != 0 {
		fmt.Printf("invalid call: req: %q, resp: %+v\n", callReq, resp)
	}
	resp.Seq = p.Seq()

	// 包装响应
	respPack := codec.CallResp2Packet(resp, p)

	// 写回给调用方
	if _, err = worker.AsyncWrite(respPack); err != nil {
		fmt.Println("write to worker failed: ", err)
	}
}
