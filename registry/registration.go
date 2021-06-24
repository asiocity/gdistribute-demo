package registry

type ServiceName string

type Registration struct {
	ServiceName ServiceName
	ServiceURL  string
	// 存放服务依赖的服务
	RequiredServices []ServiceName
	// 存放服务自己的 URL 地址, 当自己依赖用的服务发生变更, 可以让注册中心通过这个地址告知服务
	ServiceUpdateURL string
	// 用于做心跳检查
	HeartbeatURL string
}

const (
	LogService     = ServiceName("LogService")
	LibraryService = ServiceName("LibraryService")
)

type patchEntry struct {
	Name ServiceName
	URL  string
}

type patch struct {
	Added   []patchEntry
	Removed []patchEntry
}
