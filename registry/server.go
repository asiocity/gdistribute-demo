package registry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"
)

const (
	ServerPort = "3000"
	ServerURL  = "http://localhost:" + ServerPort + "/services"
)

type registry struct {
	// 目前已经注册的服务
	registrations []Registration
	// 可能会被多个协程并发访问
	lock *sync.RWMutex
}

func (r *registry) add(reg Registration) error {
	r.lock.Lock()
	r.registrations = append(r.registrations, reg)
	r.lock.Unlock()

	// 在服务注册的时候还会进行依赖服务的声明
	if err := r.sendRequiredServices(reg); err != nil {
		log.Println("send required services failed")
		return err
	}

	r.notify(patch{
		Added: []patchEntry{
			patchEntry{
				Name: reg.ServiceName,
				URL:  reg.ServiceURL,
			},
		},
	})

	return nil
}

func (r *registry) notify(fullPath patch) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	for _, reg := range r.registrations {
		// 针对每个注册的服务都开一个 goroutine
		go func(reg Registration) {
			// 对每个服务所依赖的服务进行循环
			for _, reqService := range reg.RequiredServices {
				p := patch{Added: []patchEntry{}, Removed: []patchEntry{}}
				sendUpdate := false

				for _, added := range fullPath.Added {
					if added.Name == reqService {
						p.Added = append(p.Added, added)
						sendUpdate = true
					}
				}
				for _, removed := range fullPath.Removed {
					if removed.Name == reqService {
						p.Removed = append(p.Removed, removed)
						sendUpdate = true
					}
				}
				if sendUpdate {
					if err := r.sendPatch(p, reg.ServiceUpdateURL); err != nil {
						log.Println(err)
						return
					}
				}
			}
		}(reg)
	}

}

func (r *registry) sendRequiredServices(reg Registration) error {
	r.lock.RLock()
	defer r.lock.RUnlock()

	// 有增有减的
	var p patch
	// 循环查找已注册的服务
	for _, serviceReg := range r.registrations {
		// 循环查找依赖的服务
		for _, reqService := range reg.RequiredServices {
			if serviceReg.ServiceName == reqService {
				p.Added = append(p.Added, patchEntry{
					Name: serviceReg.ServiceName,
					URL:  serviceReg.ServiceURL,
				})
			}
		}
	}
	// 通过更新 URL 把 patch / 依赖的相关服务的 URL 发送过去
	if err := r.sendPatch(p, reg.ServiceUpdateURL); err != nil {
		return err
	}

	return nil
}

func (r *registry) sendPatch(p patch, updateURL string) error {
	data, err := json.Marshal(p)
	if err != nil {
		return err
	}

	_, err = http.Post(updateURL, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	return nil
}

func (r *registry) remove(url string) error {
	for i := range r.registrations {
		if url == r.registrations[i].ServiceURL {
			r.notify(patch{
				Removed: []patchEntry{
					{
						Name: r.registrations[i].ServiceName,
						URL:  r.registrations[i].ServiceURL,
					},
				},
			})

			r.lock.Lock()
			defer r.lock.Unlock()

			r.registrations = append(r.registrations[:i], r.registrations[i+1:]...)

			return nil
		}
	}
	return fmt.Errorf("service at URL %s not found", url)
}

var reg = registry{
	registrations: make([]Registration, 0),
	lock:          new(sync.RWMutex),
}

func (r *registry) heartbeat(freq time.Duration) {
	for {
		var wg sync.WaitGroup
		for _, reg := range r.registrations {
			wg.Add(1)
			go func(reg Registration) {
				defer wg.Done()

				success := true
				for attemps := 0; attemps < 3; attemps++ {
					res, err := http.Get(reg.HeartbeatURL)
					// 请求失败
					if err != nil {
						log.Println(err)
					} else if res.StatusCode == http.StatusOK {
						// 请求成功
						log.Printf("Heartbeat check passed for %v", reg.ServiceName)
						// 判断是否有失败过, 失败过则重新把服务添加回 r 中
						if !success {
							r.add(reg)
						}
						break
					}

					// 请求失败继续往下走
					log.Printf("Heartbeat check failed for %v", reg.ServiceName)
					if success {
						success = false
						r.remove(reg.ServiceURL)
					}

					// 等 1s 重试
					time.Sleep(time.Second * 1)
				}
			}(reg)
		}
	}
}

var once sync.Once

func SetupRegistryService() {
	once.Do(func() {
		go reg.heartbeat(time.Second * 3)
	})
}

// 让它成为 HTTP Server
type RegistryService struct{}

func (s RegistryService) ServeHTTP(
	w http.ResponseWriter,
	r *http.Request,
) {
	log.Println("Request received")
	switch r.Method {
	case http.MethodPost:
		dec := json.NewDecoder(r.Body)

		var r Registration
		if err := dec.Decode(&r); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Printf("Adding service: %v with URL: %s\n", r.ServiceName, r.ServiceURL)

		if err := reg.add(r); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	case http.MethodDelete:
		payload, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		url := string(payload)
		log.Printf("Removing service at URL: %s", url)
		if err := reg.remove(url); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

}
