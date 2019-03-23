package server

import (
	"fmt"
	"log"
	"nix/codec"
	"nix/registry"
	"nix/service"
	"time"
	"tron"
)

type NixServer struct {
	server  *tron.Server
	manager *service.ServiceManager
}

func NewNixServer(addr string, services []service.Service) *NixServer {
	conf := tron.NewConfig(16*1024, 16*1024, 100, 100, 1000, 5*time.Second)
	s := &NixServer{}
	s.manager = service.NewServiceManager(services)
	s.server = tron.NewServer(addr, conf, codec.NewServerCodec(), s.packetHandler)

	if err := s.server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}

	center := registry.NewRegistryCenter(addr, services) //
	center.RegisterAllService()                          // 注册所有服务

	return s
}

// 执行本地调用
func (s NixServer) packetHandler(worker *tron.Client, p *tron.Packet) {
	callReq, err := codec.CmdReq2CallReq(p.Data)
	if err != nil {
		fmt.Println("packetHandler", err)
	}
	callReq.Timeout = 5 * time.Second

	// 执行调用
	resp := s.manager.Call(*callReq)
	if resp.Ec != 0 {
		fmt.Printf("invalid call: req: %+v, resp: %+v\n", callReq, resp)
	}
	resp.Seq = p.Seq()

	// 包装响应
	respPack := codec.CallResp2Packet(resp, p)

	// 写回给调用方
	if _, err = worker.AsyncWrite(respPack); err != nil {
		fmt.Println("write to worker failed: ", err)
	}
}
