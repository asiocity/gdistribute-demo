package service

import (
	"context"
	"distributed/registry"
	"fmt"
	"log"
	"net/http"
)

// 用来集中启动所有 service 服务
func Start(
	ctx context.Context,
	host, port string,
	reg registry.Registration,
	registerHandlersFunc func(),
) (context.Context, error) {
	// 注册 HTTP 服务
	registerHandlersFunc()

	// 启动服务
	ctx = startService(ctx, reg.ServiceName, host, port)

	// 注册服务到注册中心
	if err := registry.RegisterService(reg); err != nil {
		return ctx, err
	}

	return ctx, nil
}

func startService(
	ctx context.Context,
	serviceName registry.ServiceName,
	host, port string,
) context.Context {
	ctx, cancel := context.WithCancel(ctx)

	var srv http.Server
	srv.Addr = ":" + port

	// 启动服务
	go func() {
		// srv.ListenAndServe() 是阻塞的, 返回错误
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}

		// 退出后移除服务
		if err := registry.ShutdownService(fmt.Sprintf("http://%s:%s", host, port)); err != nil {
			log.Println(err)
		}

		// 如果是正常退出则直接退出, 如果是异常的话就执行 cancel()
		select {
		case <-ctx.Done():
			break
		default:
			cancel()
		}
	}()

	// 等待手动停止
	go func() {
		fmt.Printf("%v started. Press any key to stop.\n", serviceName)

		var s string
		fmt.Scanln(&s)

		// if err := registry.ShutdownService(fmt.Sprintf("http://%s:%s", host, port)); err != nil {
		// log.Println(err)
		// }

		srv.Shutdown(ctx)

		cancel()
	}()

	return ctx
}
