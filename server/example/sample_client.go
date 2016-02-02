package main
import (
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
	"golang.org/x/net/context"
	"github.com/maniksurtani/qs/quotaservice/protos"
)

func main() {
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
	rsp, err := client.Allow(context.TODO(), req)
	if err != nil {
		fmt.Printf("Caught error %v", err)
	} else {
		fmt.Printf("Got response %v", rsp)
	}
}

