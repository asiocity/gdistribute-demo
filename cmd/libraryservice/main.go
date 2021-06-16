package main

import (
	"context"
	"distributed/library"
	"distributed/log"
	"distributed/registry"
	"distributed/service"
	"fmt"
	stlog "log"
)

func main() {
	host, port := "localhost", "6000"
	serviceAddr := fmt.Sprintf("http://%s:%s", host, port)

	r := registry.Registration{
		ServiceName: registry.LibraryService,
		ServiceURL:  serviceAddr,
		// RequiredServices: []registry.ServiceName{
		// 	registry.LibraryService,
		// },
		// ServiceUpdateURL: serviceAddr + "/services",
	}
	ctx, err := service.Start(
		context.Background(),
		host,
		port,
		r,
		library.RegisterHandlers,
	)
	if err != nil {
		stlog.Fatalln(err)
	}

	if logProvider, err := registry.GetProvider(registry.LogService); err == nil {
		fmt.Printf("Logging service found at: %s\n", logProvider)
		log.SetClientLogger(logProvider, r.ServiceName)
	}

	// 等待停止
	<-ctx.Done()

	fmt.Println("Shutting down library service")
}
