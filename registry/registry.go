package registry

import (
	"context"
)

//接口定义  注册组件接口
type Registry interface {
	Name() string                                                 //插件名字
	Init(ctx context.Context, opts ...Option) (err error)         //初始化组件
	Register(ctx context.Context, sevice *Service) (err error)    //服务注册到etcd
	Unregister(ctx context.Context, service *Service) (err error) //反注册，从注册中心下掉
}
