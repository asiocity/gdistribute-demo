package registry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"sync"
)

// 用于给 RegistryService 发送一个 POST 请求
func RegisterService(r Registration) error {
	heartbeatURL, err := url.Parse(r.HeartbeatURL)
	if err != nil {
		return err
	}
	http.HandleFunc(heartbeatURL.Path, func(w http.ResponseWriter, r *http.Request) {
		// 如果是生产环境一般还会返回 CPU MEM 等其他信息
		w.WriteHeader(http.StatusOK)
	})

	serviceUpdateURL, err := url.Parse(r.ServiceUpdateURL)
	if err != nil {
		return err
	}
	http.Handle(serviceUpdateURL.Path, &serviceUpdateHandler{})

	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	if err := enc.Encode(r); err != nil {
		return err
	}

	res, err := http.Post(ServerURL, "application/json", buf)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to register service. Registry service"+
			"responded with code %v", res.StatusCode)
	}

	return nil
}

type serviceUpdateHandler struct{}

func (sh *serviceUpdateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 只允许 post
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	dec := json.NewDecoder(r.Body)
	var p patch
	err := dec.Decode(&p)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	fmt.Printf("Updated reviced %v\n", p)
	prov.Update(p)
}

// 用于取消服务
func ShutdownService(url string) error {
	// http 包中没有单独的 del 函数
	req, err := http.NewRequest(
		http.MethodDelete,
		ServerURL,
		bytes.NewBuffer([]byte(url)),
	)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "text/plain")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to deregister service, Registry "+
			"service responded with code %v", res.StatusCode)
	}

	return nil
}

// 被依赖的服务给其他服务使用
type providers struct {
	// 一个服务可能有多个 url 因此用 []string
	services map[ServiceName][]string
	lock     *sync.RWMutex
}

func (p *providers) Update(pat patch) {
	p.lock.Lock()
	defer p.lock.Unlock()

	// added
	for _, patchEntry := range pat.Added {
		// 如果服务名称不存在
		if _, ok := p.services[patchEntry.Name]; !ok {
			p.services[patchEntry.Name] = make([]string, 0)
		}
		p.services[patchEntry.Name] = append(p.services[patchEntry.Name], patchEntry.URL)
	}

	// removed
	for _, patchEntry := range pat.Removed {
		// 如果服务名称存在
		if providerURLs, ok := p.services[patchEntry.Name]; ok {
			for i := range providerURLs {
				if providerURLs[i] == patchEntry.URL {
					p.services[patchEntry.Name] = append(providerURLs[:i], providerURLs[i+1:]...)
				}
			}
		}
	}
}

// 使用服务名称来找到它的 URL
// 偷懒, 本来该返回 []string
func (p providers) get(name ServiceName) (string, error) {
	providers, ok := p.services[name]
	if !ok {
		return "", fmt.Errorf("no providers avaliable for service %v", name)
	}

	// 偷懒, 本来应该返回 []string 的
	idx := int(rand.Float32() * float32(len(providers)))
	return providers[idx], nil
}

func GetProvider(name ServiceName) (string, error) {
	return prov.get(name)
}

var prov = providers{
	services: make(map[ServiceName][]string),
	lock:     new(sync.RWMutex),
}
