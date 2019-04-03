package tracing

import (
	"fmt"

	"github.com/opentracing/opentracing-go/log"
)

func ErrorField(err error) log.Field {
	//为什么不使用log.Error？
	//因为它最终输出会是err.Error()，不会将pkg.error的stacks打印出来
	//我们一些时候还是希望有这种信息
	//jaeger里对object类型的filed会做fmt.Sprintf("%+v",obj)处理
	//所以我们用这个
	return log.Object("error", err)
}

func PruneBodyLog(log string, maxBodyLogSize int) string {
	if maxBodyLogSize <= 0 { //无需修剪
		return log
	}

	le := len(log)
	if le <= maxBodyLogSize {
		return log
	}

	r := fmt.Sprintf("Body is too large(%d), just prune to %d-->\n", le, maxBodyLogSize)
	return r + log[0:maxBodyLogSize-len(r)]
}
