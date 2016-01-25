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
	serverAddr := "127.0.0.1:11111"
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	conn, err := grpc.Dial(serverAddr, opts...)
	if err != nil {
		grpclog.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()

	client := quotaservice.NewQuotaServiceClient(conn)
	rq := int32(1024)
	svcA := "A"
	svcB := "B"
	req := &quotaservice.GetQuotaRequest{NumTokensRequested: &rq, FromService:&svcA, ToService:&svcB}
	rsp, err := client.GetQuota(context.TODO(), req)
	if err != nil {
		fmt.Printf("Caught error %v", err)
	} else {
		fmt.Printf("Got response %v", rsp)
	}
}

