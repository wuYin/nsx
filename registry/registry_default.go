package registry

import (
	"encoding/json"
	"fmt"
	"gopkg.in/redis.v3"
	"nix/codec"
	"nix/service"
)

type DefaultRegistry struct {
	addr   string
	regCli *redis.Client
}

func NewDefaultRegistry(addr string) *DefaultRegistry {
	cli := redis.NewClient(&redis.Options{
		Addr:     addr,
		DB:       0,
		Password: "",
	})
	return &DefaultRegistry{addr: addr, regCli: cli}
}

// 向 admin-service 发起注册请求
func (r DefaultRegistry) RegisterService(newServiceUri string) error {
	cmd := codec.CmdReq{
		ServiceUri: service.SERVICE_ADMIN,
		Method:     service.METHOD_REGISTER,
		Args: []interface{}{
			newServiceUri,
		},
	}
	_, err := r.request(cmd)
	return err
}

// 将服务下线
func (r DefaultRegistry) UnRegisterService(uri string) error {
	cmd := codec.CmdReq{
		ServiceUri: service.SERVICE_ADMIN,
		Method:     service.METHOD_UNREGISTER,
		Args:       []interface{}{uri},
	}
	_, err := r.request(cmd)
	return err
}

// 获取服务准备调用
func (r DefaultRegistry) GetService(uri string) (addr string) {
	cmd := codec.CmdReq{
		ServiceUri: service.SERVICE_ADMIN,
		Method:     service.METHOD_GET_SERVICE,
		Args:       []interface{}{uri},
	}

	resp, err := r.request(cmd)
	if err != nil {
		return ""
	}
	s, ok := resp.(string)
	if !ok {
		return ""
	}

	return s
}

type RegisterResp struct {
	Ec  int         `json:"ec"`
	Em  string      `json:"em"`
	Res interface{} `json:"res"`
}

// 将 cmd 直接序列化为 json 字符串发起 GET 调用
func (r DefaultRegistry) request(cmd codec.CmdReq) (interface{}, error) {
	b, _ := json.Marshal(cmd)
	s, err := r.regCli.Get(string(b)).Result()

	if err != nil {
		fmt.Printf("codec request failed: %v\n", err)
		return "", err
	}

	var resp RegisterResp
	if err = json.Unmarshal([]byte(s), &resp); err != nil {
		fmt.Printf("codec unmarshal to resp failed: %v\n", err)
		return "", err
	}

	return resp.Res, nil
}
