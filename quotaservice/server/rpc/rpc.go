package rpc
import (
	"github.com/maniksurtani/qs/quotaservice/server/configs"
	"github.com/maniksurtani/qs/quotaservice/server/service"
)

type RpcEndpoint interface {
	Init(cfgs *configs.Configs, qs service.QuotaService)
	Start()
	Stop()
}
