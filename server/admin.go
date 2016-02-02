package server

import (
	"fmt"
	"net/http"
	"github.com/maniksurtani/qs/quotaservice/server/logging"
)

type AdminServer struct {
	port int
}

func NewAdminServer(port int) *AdminServer {
	s := AdminServer{port: port}
	return &s
}

func (this *AdminServer) Start() {
	logging.Printf("Starting admin console on port %v", this.port)
	http.HandleFunc("/", handler)
	go http.ListenAndServe(fmt.Sprintf(":%v", this.port), nil)
}

func (this *AdminServer) Stop() {
	logging.Print("Stopping admin console")
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "A future admin console")
}

