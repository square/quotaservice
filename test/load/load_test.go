package loadtest
import (
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
	"golang.org/x/net/context"
	"testing"
	"github.com/maniksurtani/quotaservice/protos"
)

func BenchmarkQuotaRequests(b *testing.B) {
	fmt.Println("Starting example client.")
	serverAddr := "127.0.0.1:10990"
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	conn, err := grpc.Dial(serverAddr, opts...)
	if err != nil {
		grpclog.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()

	client := quotaservice.NewQuotaServiceClient(conn)

	req := &quotaservice.AllowRequest{
		BucketName: "one",
		TokensRequested: 1,
	}

	b.ResetTimer()
	b.SetParallelism(8)
	b.RunParallel(
		func(pb *testing.PB) {
			for pb.Next() {
				client.Allow(context.TODO(), req)
			}
		})
}

