// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/maniksurtani/quotaservice/master/LICENSE

package loadtest

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	pb "github.com/maniksurtani/quotaservice/protos"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
	"testing"
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

	client := pb.NewQuotaServiceClient(conn)

	req := &pb.AllowRequest{
		Namespace:          proto.String("test.namespace"),
		Name:               proto.String("one"),
		NumTokensRequested: proto.Int64(1)}
	b.ResetTimer()
	b.SetParallelism(8)
	b.RunParallel(
		func(pb *testing.PB) {
			for pb.Next() {
				client.Allow(context.TODO(), req)
			}
		})
}
