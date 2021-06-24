package main

import (
	"context"
	"distributed/log"
	"distributed/registry"
	"distributed/service"
	"fmt"
	stlog "log"
)

func main() {
	log.Run("./distributed.log")

	var (
		host        = "localhost"
		port        = "4000"
		serviceAddr = fmt.Sprintf("http://%s:%s", host, port)
		serviceName = registry.LogService
	)

	r := registry.Registration{
		ServiceName: registry.ServiceName(serviceName),
		ServiceURL:  serviceAddr,
		// 依赖一个空服务
		RequiredServices: make([]registry.ServiceName, 0),
		ServiceUpdateURL: serviceAddr + "/services",
		HeartbeatURL:     serviceAddr + "/heartbeat",
	}

	ctx, err := service.Start(
		context.Background(),
		host,
		port,
		r,
		log.RegisterHandlers,
	)

	if err != nil {
		stlog.Fatalln(err)
	}
	// 等待停止
	<-ctx.Done()

	fmt.Println("Shutting down log service")
}
