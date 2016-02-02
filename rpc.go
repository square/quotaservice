package quotaservice
import (
	"github.com/maniksurtani/quotaservice/configs"
)

type RpcEndpoint interface {
	Init(cfgs *configs.Configs, qs QuotaService)
	Start()
	Stop()
}
