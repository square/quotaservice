package main
import (
	"github.com/maniksurtani/qs/quotaservice/server"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	server := server.New("/tmp/test.yaml")
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

