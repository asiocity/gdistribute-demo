package registry

type ServiceName string

type Registration struct {
	ServiceName ServiceName
	ServiceURL  string
	// RequiredServices []ServiceName // 存放服务依赖的服务
	// ServiceUpdateURL string        // 用于和服务注册中心沟通
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
	Added []patchEntry
	Deled []patchEntry
}
