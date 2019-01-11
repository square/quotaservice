// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/square/quotaservice/master/LICENSE

package loadtest

import (
	"fmt"
	"testing"
	"context"

	pb "github.com/square/quotaservice/protos"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
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
	defer func() { _ = conn.Close() }()

	client := pb.NewQuotaServiceClient(conn)

	req := &pb.AllowRequest{
		Namespace:       "test.namespace",
		BucketName:      "one",
		TokensRequested: 1}
	b.ResetTimer()
	b.SetParallelism(8)
	b.RunParallel(
		func(pb *testing.PB) {
			for pb.Next() {
				if _, e := client.Allow(context.TODO(), req); e != nil {
					panic(e)
				}
			}
		})
}
