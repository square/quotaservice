package server

import (
	"fmt"
	"net"
	"google.golang.org/grpc"
	"github.com/maniksurtani/qs/quotaservice/protos"
	"golang.org/x/net/context"
	"log"
	"os"
	"google.golang.org/grpc/grpclog"
	"github.com/derekparker/delve/vendor/gopkg.in/yaml.v2"
)

var logger = log.New(os.Stdout, "quotaservice: ", log.Ldate | log.Ltime | log.LUTC | log.Lshortfile)

// Configs
type Configs struct {
	Port              int
	DefaultRefillRate int `yaml:"default_refill_rate,omitempty"`
	RefillRates       map[string]map[string]int `yaml:"refill_rates,flow"`
}

// The Server type
type Server struct {
	cfgs          *Configs
	grpcServer    *grpc.Server
	currentStatus status
	stopper       *chan int
}

func (this *Server) String() string {
	return fmt.Sprintf("Quota Server running on port %v with status %v", this.cfgs.Port, this.currentStatus)
}

func (this *Server) Start() (bool, error) {
	// Set up buckets
	for from, conns := range this.cfgs.RefillRates {
		for to, rate := range conns {
			b := NewBucket(bucketName(from, to), rate)
			b.Start()
		}
	}

	// Start ticker
	StartTicker()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", this.cfgs.Port))
	if err == nil {
		grpclog.SetLogger(logger)
		this.grpcServer = grpc.NewServer()
		// Each service should be registered
		quotaservice.RegisterQuotaServiceServer(this.grpcServer, this)
		go this.grpcServer.Serve(lis)
		this.currentStatus = started
		logger.Printf("Starting server on port %v", this.cfgs.Port)
		logger.Printf("Server status: %v", this.currentStatus)
		return true, nil
	} else {
		// TODO(manik) Panic?
		logger.Fatalf("Cannot start server on port %v. Error %v", this.cfgs.Port, err)
		return false, err
	}
}

func (this *Server) Stop() (bool, error) {
	// Stop ticker
	StopTicker()
	this.currentStatus = stopped
	logger.Printf("Stopping server on port %v.", this.cfgs.Port)
	this.grpcServer.Stop()
	return true, nil
}

func (this *Server) GetQuota(ctx context.Context, req *quotaservice.GetQuotaRequest) (*quotaservice.GetQuotaResponse, error) {
	logger.Printf("Handling request %v", req)
	rsp := new(quotaservice.GetQuotaResponse)
	b := BucketRegistry[bucketName(*req.FromService, *req.ToService)]
	if b == nil {
		return nil, fmt.Errorf("Unable to find bucket %v", bucketName)
	}
	granted := int32(b.Acquire(int(*req.NumTokensRequested)))
	rsp.NumTokensGranted = &granted
	return rsp, nil
}

// Constructor
func New(cfgs string) *Server {
	file, err := os.Open(cfgs)
	defer file.Close()
	if err != nil {
		// TODO(manik) panic?
		log.Fatalf("Unable to open config file %v. Error: %v", cfgs, err)
		return nil
	} else {
		// TODO(manik) is this the right approach to fully parse a YAML file?
		buf := make([]byte, 1024, 1024)
		var cfgs Configs
		cfgStream := make([]byte, 0)
		bytesRead, eof := file.Read(buf)
		for eof == nil {
			cfgStream = append(cfgStream, buf[:bytesRead]...)
			bytesRead, eof = file.Read(buf)
		}

		yaml.Unmarshal(cfgStream, &cfgs)
		return &Server{cfgs: &cfgs}
	}
}

func bucketName(from, to string) string {
	return fmt.Sprintf("%v->%v", from, to)
}

