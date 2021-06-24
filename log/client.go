package log

import (
	"bytes"
	"distributed/registry"
	"fmt"
	stlog "log"
	"net/http"
)

func SetClientLogger(serviceURL string, clientService registry.ServiceName) {
	// 先写日志
	stlog.SetPrefix(fmt.Sprintf("[%v] - ", clientService))
	// 服务端设置了时间戳, 客户端这边不用带了
	stlog.SetFlags(0)
	// 把日志发送到服务端
	stlog.SetOutput(&clientLogger{
		url: serviceURL,
	})
}

// 需要实现 io.Write 接口
type clientLogger struct {
	url string
}

func (cl clientLogger) Write(data []byte) (int, error) {
	b := bytes.NewBuffer([]byte(data))
	// 通过 post 请求将日志发送给服务端
	res, err := http.Post(cl.url+"/log", "text/plain", b)
	if err != nil {
		return 0, err
	}

	if res.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("failed to send log message. Service responded %d", res.StatusCode)
	}

	// 如果没有问题返回发送的数据长度
	return len(data), nil
}
