package registry

import (
	"time"
)

type Options struct {
	Addrs   []string      //注册中心地址
	Timeout time.Duration //跟注册中心交互的超时
	//xxx_company/app/kuaishou/service_A/10.192.1.1:8801
	//key
	RegistryPath string //注册路径，在etc里面
	HeartBeat    int64  //心跳时间
}

type Option = func(opts *Options)

func WithTimeout(timeout time.Duration) Option { //初始化timeout选项
	return func(opts *Options) {
		opts.Timeout = timeout
	}
}

func WithAddrs(addrs []string) Option { //初始化地址选项
	return func(opts *Options) {
		opts.Addrs = addrs
	}
}

func WithRegistryPath(path string) Option {
	return func(opts *Options) {
		opts.RegistryPath = path
	}
}

func WithHeartBeat(heartHeat int64) Option {
	return func(opts *Options) {
		opts.HeartBeat = heartHeat
	}
}
