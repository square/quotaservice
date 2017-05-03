package client

import (
	"math"
	"os"
	"testing"
	"time"

	"github.com/maniksurtani/quotaservice"
	"github.com/maniksurtani/quotaservice/buckets/memory"
	"github.com/maniksurtani/quotaservice/config"
	pb "github.com/maniksurtani/quotaservice/protos"
	qsgrpc "github.com/maniksurtani/quotaservice/rpc/grpc"
	"github.com/maniksurtani/quotaservice/test/helpers"
	"google.golang.org/grpc"
)

const target = "localhost:10990"

var server quotaservice.Server

func TestMain(m *testing.M) {
	setUp()
	r := m.Run()
	_, _ = server.Stop()
	os.Exit(r)
}

func setUp() {
	// Start a QuotaService server.
	cfg := config.NewDefaultServiceConfig()
	cfg.GlobalDefaultBucket = config.NewDefaultBucketConfig(config.DefaultBucketName)
	nsc := config.NewDefaultNamespaceConfig("delaying")
	bc := config.NewDefaultBucketConfig("delaying")
	bc.Size = 1 // Very small.
	bc.FillRate = 100

	helpers.PanicError(config.AddBucket(nsc, bc))
	helpers.PanicError(config.AddNamespace(cfg, nsc))

	server = quotaservice.New(memory.NewBucketFactory(),
		config.NewMemoryConfig(cfg),
		quotaservice.NewReaperConfigForTests(),
		0,
		qsgrpc.New(target))

	if _, err := server.Start(); err != nil {
		helpers.PanicError(err)
	}
}

func TestClient(t *testing.T) {
	client, err := New(target, grpc.WithInsecure())
	helpers.CheckError(t, err)

	req := &pb.AllowRequest{
		Namespace:             "Doesn't exist",
		BucketName:            "Doesn't exist",
		TokensRequested:       1,
		MaxWaitMillisOverride: math.MaxInt64}
	resp, err := client.Allow(req)
	helpers.CheckError(t, err)
	if resp.Status != pb.AllowResponse_OK {
		t.Fatalf("Expected OK. Was %v", pb.AllowResponse_Status_name[int32(resp.Status)])
	}
}

func TestBlockingClient(t *testing.T) {
	// Claim tokens
	client, err := New(target, grpc.WithInsecure())
	helpers.CheckError(t, err)

	waitTime := 0
	n := 0
	req := &pb.AllowRequest{
		Namespace:             "delaying",
		BucketName:            "delaying",
		TokensRequested:       1,
		MaxWaitMillisOverride: math.MaxInt64}

	// Consume all readily available tokens
	for waitTime == 0 {
		resp, err := client.Allow(req)
		helpers.CheckError(t, err)
		if resp.Status != pb.AllowResponse_OK {
			t.Fatalf("Expected OK. Was %v\n", pb.AllowResponse_Status_name[int32(resp.Status)])
		}
		waitTime = int(resp.WaitMillis)
		n++
	}

	// Next attempt should block at least waitTime millis.
	start := time.Now()
	e := client.AllowBlocking(req)
	elapsed := time.Since(start)
	helpers.CheckError(t, e)
	expected := time.Duration(waitTime) * time.Millisecond

	if elapsed < expected {
		t.Fatalf("Expected to block for at least %v but was just %v", expected, elapsed)
	}

}
