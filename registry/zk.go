package registry

import (
	"fmt"
	"github.com/samuel/go-zookeeper/zk"
	"strings"
	"time"
)

type zkRegistry struct {
	conn *zk.Conn
}

const ZK_ROOT = "/services"

// zk 注册代理
func NewZKRegistry(zkServers []string) *zkRegistry {
	conn, _, err := zk.Connect(zkServers, 10*time.Second)
	if err != nil {
		panic(err)
	}

	exist, _, err := conn.Exists(ZK_ROOT)
	if err != nil {
		panic(err)
	}

	if !exist {
		if _, err = conn.Create(ZK_ROOT, []byte(""), 0, zk.WorldACL(zk.PermAll)); err != nil {
			panic(err)
		}
	}
	return &zkRegistry{conn: conn}
}

// 服务注册
func (r *zkRegistry) Register(uri, addr string) error {
	path := genPath(uri, addr)
	for _, path := range getSubPaths(path) {
		exist, _, err := r.conn.Exists(path)
		if err != nil {
			fmt.Println("path", path)
			return err
		}
		if !exist {
			if _, err = r.conn.Create(path, []byte(""), 0, zk.WorldACL(zk.PermAll)); err != nil {
				return err
			}
		}
	}
	return nil
}

// 服务下线
func (r *zkRegistry) UnRegister(uri, addr string) error {
	path := genPath(uri, addr)
	exist, _, err := r.conn.Exists(path)
	if err != nil {
		return err
	}
	if exist {
		if err = r.conn.Delete(path, 0); err != nil {
			return err
		}
	}
	return nil
}

// 服务发现
func (r *zkRegistry) GetService(uri string) ([]string, error) {
	path := genPath(uri)
	children, _, err := r.conn.Children(path)
	if err != nil {
		return nil, err
	}
	return children, nil
}

func genPath(uri string, addrs ...string) string {
	if len(addrs) == 0 {
		return ZK_ROOT + "/" + uri
	}
	return ZK_ROOT + "/" + uri + "/" + addrs[0]
}

func getSubPaths(path string) (paths []string) {
	subPaths := strings.Split(path, "/")
	var s string
	for _, p := range subPaths[1:] {
		paths = append(paths, s+"/"+p)
		s += "/" + p
	}
	return
}
