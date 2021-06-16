package registry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
)

const (
	ServerPort = "3000"
	ServerURL  = "http://localhost:" + ServerPort + "/services"
)

type registry struct {
	// 目前已经注册的服务
	// 可能会被多个协程并发访问
	registrations []Registration
	lock          *sync.RWMutex
}

func (r *registry) add(reg Registration) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.registrations = append(r.registrations, reg)

	// 在服务注册的时候还会进行依赖服务的声明
	// err := r.sendRequiredServices(reg)
	// _ = r.sendRequiredServices(reg)

	return nil
}

// func (r *registry) sendRequiredServices(reg Registration) error {
// 	r.lock.RLock()
// 	defer r.lock.RUnlock()
//
// 	// 有增有减的
// 	var p patch
// 	// 循环查找已注册的服务
// 	for _, serviceReg := range r.registrations {
// 		// 循环查找依赖的服务
// 		for _, reqService := range reg.RequiredServices {
// 			if serviceReg.ServiceName == reqService {
// 				p.Added = append(p.Added, patchEntry{
// 					Name: serviceReg.ServiceName,
// 					URL:  serviceReg.ServiceURL,
// 				})
// 			}
// 		}
// 	}
// 	// 通过更新 URL 把 patch / 依赖的相关服务的 URL 发送过去
// 	if err := r.sendPatch(p, reg.ServiceUpdateURL); err != nil {
// 		return err
// 	}
//
// 	return nil
// }

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

func (r *registry) del(url string) error {
	for i := range r.registrations {
		if url == r.registrations[i].ServiceURL {
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
		if err := reg.del(url); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

}
