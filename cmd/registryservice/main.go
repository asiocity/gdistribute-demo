package main

import (
	"context"
	"distributed/registry"
	"fmt"
	"log"
	"net/http"
	"sync"
)

func main() {
	registry.SetupRegistryService()
	http.Handle("/services", &registry.RegistryService{})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var srv http.Server
	srv.Addr = ":" + registry.ServerPort

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		log.Println(srv.ListenAndServe())
		cancel()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		fmt.Println("Registry service started, Press any key to stop.")

		var s string
		fmt.Scanln(&s)

		srv.Shutdown(ctx)
		cancel()
	}()

	<-ctx.Done()
	fmt.Println("Shutting down registry service")

	wg.Wait()
}
