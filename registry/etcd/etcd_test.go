package etcd

import (
	"context"
	"testing"
	"time"

	"registry"
)

func TestRegister(t *testing.T) {
	registryInst, err := registry.InitRegistry(context.TODO(), "etcd",
		registry.WithAddrs([]string{"192.168.249.128:2379"}),
		registry.WithTimeout(time.Second),
		registry.WithRegistryPath("/ibinarytree/koala/"),
		//registry.WithRegistryPath("C:/test/abc/"),
		registry.WithHeartBeat(5), //5秒一个调度
	)
	if err != nil {
		t.Errorf("init registry failed, err:%v", err)
		return
	}

	service := &registry.Service{
		Name: "comment_service",
	}

	service.Nodes = append(service.Nodes, &registry.Node{
		IP:   "127.0.0.1",
		Port: 8801,
	})
	registryInst.Register(context.TODO(), service)
	for {
		time.Sleep(time.Second)
	}

}
