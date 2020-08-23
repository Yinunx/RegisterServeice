package registry

//数据结构设计

//服务抽象
type Service struct {
	Name  string  `json:"name"`
	Nodes []*Node `json:"nodes"`
}

//服务节点的抽象
type Node struct {
	Id   string `json:"id"`
	IP   string `json:"ip"`
	Port int    `json:"port"`
}
