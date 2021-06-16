// 日志服务, 接收 POST 请求, 把请求中的内容写到 LOG 文件中

package log

import (
	"io/ioutil"
	stlog "log"
	"net/http"
	"os"
)

var log *stlog.Logger

type fileLog string
type aLog string

func (fl fileLog) Write(data []byte) (int, error) {
	f, err := os.OpenFile(string(fl), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	return f.Write(data)
}

func Run(dst string) {
	// fileLog 就是 io.Write 类型, 因为定义了 Write()
	// LstdFlags 是标准 Logger 的初始值, Ldata | Ltime 组成
	log = stlog.New(fileLog(dst), "destributed - ", stlog.LstdFlags)
}

func RegisterHandlers() {
	http.HandleFunc("/log", func(rw http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			msg, err := ioutil.ReadAll(r.Body)
			if err != nil || len(msg) == 0 {
				rw.WriteHeader(http.StatusBadRequest)
				return
			}
			write(string(msg))
		default:
			rw.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	})
}

func write(msg string) {
	log.Printf("%v\n", msg)
}
