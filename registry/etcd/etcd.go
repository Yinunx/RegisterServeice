package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"time"

	"registry"

	"github.com/coreos/etcd/clientv3"
)

//服务注册插件接口
/*
type Registry interface {
	Name() string                                                 //插件名字
	Init(opts ...Option) (err error)                              //初始化组件
	Register(ctx context.Context, sevice *Service) (err error)    //服务注册到etcd
	Unregister(ctx context.Context, service *Service) (err error) //反注册，从注册中心下掉
}
*/

const (
	MaxServiceNum = 8
)

type EtcdRegistry struct {
	options   *registry.Options
	client    *clientv3.Client
	serviceCh chan *registry.Service //注册放在channel里面

	registryServiceMap map[string]*RegisterService
}

var (
	etcdRegistry *EtcdRegistry = &EtcdRegistry{
		serviceCh:          make(chan *registry.Service, MaxServiceNum),
		registryServiceMap: make(map[string]*RegisterService, MaxServiceNum),
	}
)

type RegisterService struct {
	id         clientv3.LeaseID //租约id
	service    *registry.Service
	registered bool //有没有被注册

	keepAliveCh <-chan *clientv3.LeaseKeepAliveResponse //管道
}

func init() {
	registry.RegisterPlugin(etcdRegistry)
	go etcdRegistry.run() //注册到etcd中去
}

//插件的名字
func (e *EtcdRegistry) Name() string {
	return "etcd"
}

//初始化
func (e *EtcdRegistry) Init(ctx context.Context, opts ...registry.Option) (err error) {
	e.options = &registry.Options{}
	for _, opt := range opts {
		opt(e.options)
	}

	e.client, err = clientv3.New(clientv3.Config{
		Endpoints:   e.options.Addrs,
		DialTimeout: e.options.Timeout,
	}) //etcd交互

	if err != nil {
		fmt.Printf("failed init....")
		return
	}

	return
}

//服务注册
func (e *EtcdRegistry) Register(ctx context.Context, service *registry.Service) (err error) {

	select {
	case e.serviceCh <- service:
	default: //满了
		err = fmt.Errorf("register chan is full")
		return
	}
	return
}

//服务反注册
func (e *EtcdRegistry) Unregister(ctx context.Context, service *registry.Service) (err error) {
	return
}

func (e *EtcdRegistry) run() {
	for {
		//获取当前需要注册的服务
		select {
		case service := <-e.serviceCh: //管道里面有服务的话
			_, ok := e.registryServiceMap[service.Name] //如果存在
			if ok {
				break
			}
			registryService := &RegisterService{
				service: service,
			}
			e.registryServiceMap[service.Name] = registryService
		default:
			e.registerOrKeepAlive()            //续约 那些维持心跳，哪些需要注册
			time.Sleep(time.Millisecond * 500) //防止死循环
		}
	}

}

//续约
func (e *EtcdRegistry) registerOrKeepAlive() {
	for _, registryService := range e.registryServiceMap {
		if registryService.registered {
			e.KeepAlive(registryService) //已经注册，续约逻辑
			continue
		}
		e.registerService(registryService) //注册
	}

}

//续约失败，在这里处理
func (e *EtcdRegistry) KeepAlive(registerService *RegisterService) { //服务存活看管道有没有数据

	select {
	case resp := <-registerService.keepAliveCh:
		if resp == nil { //管道里面没有
			registerService.registered = false
			return
		}
		fmt.Printf("service:%s node:%s ttl:%v\n", registerService.service.Name, registerService.service.Nodes[0].IP, registerService.service.Nodes[0].Port)
	}
	return
}

func (e *EtcdRegistry) registerService(registerService *RegisterService) (err error) {

	resp, err := e.client.Grant(context.TODO(), e.options.HeartBeat) //获取租约。注册
	if err != nil {
		return
	}

	//租约的id
	registerService.id = resp.ID
	//把key拼出来，遍历多个节点
	for _, node := range registerService.service.Nodes {
		tmp := &registry.Service{
			Name: registerService.service.Name,
			Nodes: []*registry.Node{
				node,
			},
		}

		data, err := json.Marshal(tmp)
		if err != nil {
			continue
		}

		key := e.serviceNodePath(tmp)

		_, err = e.client.Put(context.TODO(), key, string(data), clientv3.WithLease(resp.ID))
		fmt.Println(key)
		if err != nil {
			continue
		}

		ch, err := e.client.KeepAlive(context.TODO(), resp.ID) //续约
		if err != nil {                                        //失败的话下次再注册
			continue
		}
		registerService.keepAliveCh = ch
		registerService.registered = true
	}

	return
}

func (e *EtcdRegistry) serviceNodePath(service *registry.Service) string {
	nodeIP := fmt.Sprintf("%s:%d", service.Nodes[0].IP, service.Nodes[0].Port)
	return path.Join(e.options.RegistryPath, service.Name, nodeIP)
}
