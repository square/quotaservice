// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/square/quotaservice/master/LICENSE

package main

import (
	"fmt"

	pb "github.com/square/quotaservice/protos"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
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
	defer func() { _ = conn.Close() }()

	client := pb.NewQuotaServiceClient(conn)

	req := &pb.AllowRequest{
		Namespace:       "test.namespace",
		BucketName:      "abc",
		TokensRequested: 1}
	rsp, err := client.Allow(context.TODO(), req)
	if err != nil {
		fmt.Printf("Caught error %v", err)
	} else {
		fmt.Printf("Got response %v", rsp)
	}
}
