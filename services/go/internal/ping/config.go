package ping

import "github.com/zeromicro/go-zero/zrpc"

// Config сервиса: zrpc.RpcServerConf даёт ListenOn, Mode, Log, Prometheus, Telemetry, Health и т.д.
type Config struct {
	zrpc.RpcServerConf
}
