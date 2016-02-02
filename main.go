package main
import (
	"github.com/maniksurtani/qs/quotaservice/server"
	"os"
	"os/signal"
	"syscall"
	"github.com/maniksurtani/qs/quotaservice/server/rpc/grpc"
)

type Clustering interface {
	// Returns true if the current node is the leader
	IsLeader() bool
	// Returns a channel that is used to notify the query service of a membership change.
	MembershipChangeNotificationChannel() chan bool
}

type MetricsHandler interface {
	// Returns true if the current node is the leader
	IsLeader() bool
	// Returns a channel that is used to notify the query service of a membership change.
	MembershipChangeNotificationChannel() chan bool
}


func main() {
	server := server.New("/tmp/test.yaml", &grpc.GrpcEndpoint{})
	server.Start()

	// Block until SIGTERM, SIGKILL or SIGINT
	sigs := make(chan os.Signal, 1)
	shutdown := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGINT)
	go func() {
		<- sigs
		shutdown <- true
	}()
	<- shutdown
	server.Stop()
}

