package service

import (
	"encoding/json"
	"fmt"
	"nsx/codec"
	"reflect"
	"time"
)

const (
	// 管理注册中心的 admin 服务
	SERVICE_ADMIN = "admin-service"

	// 方法名
	METHOD_REGISTER    = "RegisterService"
	METHOD_UNREGISTER  = "UnRegisterService"
	METHOD_GET_SERVICE = "GetService"
)

type Service struct {
	Uri       string            // 服务名称
	Instance  interface{}       // 已实现对外接口的实例
	methods   map[string]Method // 对外可调用的方法列表
	Interface reflect.Type      // 接口定义
}

type Method struct {
	Name      string
	Value     reflect.Value
	ArgTypes  []reflect.Type // 入参
	ReplyType reflect.Type   // 出参
}

type ServiceManager struct {
	uri2Service map[string]Service
}

func NewServiceManager(services []Service) *ServiceManager {
	manager := &ServiceManager{uri2Service: make(map[string]Service)}

	for _, s := range services {
		iv := reflect.ValueOf(s.Instance)
		it := reflect.TypeOf(s.Instance)
		// Implements 判断会区分 receiver 是指针还是值，调用时需统一
		if ok := it.Implements(s.Interface); !ok {
			msg := fmt.Sprintf("%s not implement interface", s.Uri)
			panic(msg)
		}

		methods := make(map[string]Method)
		for i := 0; i < it.NumMethod(); i++ {
			// 这里比较坑啊，普通类型的 Method() 和直接对该方法进行 reflect.Type()，取出来的类型是不一致的，前者第一个入参是结构类型
			// 前者是方法变量，后者是方法。值得区分
			m := s.Interface.Method(i)
			mt := m.Type
			method := Method{Name: m.Name, Value: iv.MethodByName(m.Name)}
			// 返回参数约定只有一个，比如 error
			if mt.NumOut() != 1 {
				msg := fmt.Sprintf("panic: %s %s has %d replies args", s.Uri, mt.Name(), mt.NumOut())
				panic(msg)
			}
			method.ReplyType = mt.Out(0)

			// 输入参数可多个
			for i := 0; i < mt.NumIn(); i++ {
				method.ArgTypes = append(method.ArgTypes, mt.In(i))
			}
			methods[method.Name] = method
		}
		s.methods = methods
		manager.uri2Service[s.Uri] = s
	}

	return manager
}

// 执行调用
func (m *ServiceManager) Call(req codec.CallReq) (resp codec.CallResp) {
	// 检查服务
	service, ok := m.uri2Service[req.ServiceUri]
	if !ok {
		resp.Ec = 1
		resp.Em = fmt.Sprintf("call: %s not registed", req.ServiceUri)
		return
	}

	// 检查调用方法
	method, ok := service.methods[req.Method]
	if !ok {
		resp.Ec = 2
		resp.Em = fmt.Sprintf("call: %s method %s not found", req.ServiceUri, req.Method)
		return
	}

	// 检查参数个数
	if len(req.Args) != len(method.ArgTypes) {
		resp.Ec = 3
		resp.Em = fmt.Sprintf("call: %s %s args not match: %v", req.ServiceUri, req.Method, req.Args)
		return
	}

	// 取出指定类型的参数，反序列化后直接赋值
	in := make([]reflect.Value, 0, len(req.Args))
	for i, argType := range method.ArgTypes {
		zeroVPtr := reflect.New(argType)
		if err := json.Unmarshal(req.Args[i], zeroVPtr.Interface()); err != nil {
			panic("call: unmarshal: failed" + err.Error())
		}
		in = append(in, zeroVPtr.Elem())
	}

	resCh := make(chan []reflect.Value, 1)
	go func() {
		defer func() {
			if err := recover(); err != nil {
				fmt.Println("call fatal:", err)
				resCh <- nil
			}
			resCh <- method.Value.Call(in)
		}()
	}()

	// 等待执行完毕
	select {
	case <-time.After(req.Timeout):
		resp.Ec = 4
	case res := <-resCh:
		if res == nil {
			resp.Ec = 5
			return
		}
		buf, _ := json.Marshal(res[0].Interface())
		resp.Res = buf
	}

	return resp
}
