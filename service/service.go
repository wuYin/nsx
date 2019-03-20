package service

import (
	"fmt"
	"logx"
	"reflect"
	"time"
	"velar"
)

type Service struct {
	Uri       string            // 服务名称
	Proxy     interface{}       // 已实现对外接口的代理实例
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
	services map[string]Service
}

func NewServiceManager(services []Service) *ServiceManager {
	manager := &ServiceManager{services: make(map[string]Service)}

	for _, s := range services {
		pv := reflect.ValueOf(s.Proxy)
		pt := reflect.TypeOf(s.Proxy)
		// Implements 判断会区分 receiver 是指针还是值，调用时需统一
		if ok := pt.Implements(s.Interface); !ok {
			msg := fmt.Sprintf("%s not implement interface", s.Uri)
			panic(msg)
		}

		methods := make(map[string]Method)
		for i := 0; i < pt.NumMethod(); i++ {
			// 这里比较坑啊，普通类型的 Method() 和直接对该方法进行 reflect.Type()，取出来的类型是不一致的，前者第一个入参是结构类型
			// 前者是方法变量，后者是方法。值得区分
			m := s.Interface.Method(i)
			mt := m.Type
			method := Method{Name: m.Name, Value: pv.MethodByName(m.Name)}
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
		manager.services[s.Uri] = s
	}

	return manager
}

// 执行调用
func (m *ServiceManager) Call(req velar.CallReq) (resp velar.CallResp) {
	// 检查服务
	service, ok := m.services[req.ServiceUri]
	if !ok {
		resp.Em = fmt.Sprintf("call: %s not registed", req.ServiceUri)
		return
	}

	// 检查调用方法
	method, ok := service.methods[req.Method]
	if !ok {
		resp.Ec = 2
		return
	}

	// 检查参数个数
	if len(req.Args) != len(method.ArgTypes) {
		resp.Ec = 3
		resp.Em = fmt.Sprintf("args are %d, should be %d", len(req.Args), len(method.ArgTypes))
		return
	}

	// 值转换
	refVals := make([]reflect.Value, 0, len(req.Args))
	for i := range method.ArgTypes {
		v := reflect.ValueOf(req.Args[i])
		refVals = append(refVals, v)
	}

	resCh := make(chan []reflect.Value, 1)
	go func() {
		defer func() {
			if err := recover(); err != nil {
				logx.Error(err)
				resCh <- nil
			}
			resCh <- method.Value.Call(refVals)
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
		resp.Res = res[0].Interface()
	}

	return resp
}
