// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/maniksurtani/quotaservice/master/LICENSE

package main

import (
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
	"golang.org/x/net/context"
	qspb "github.com/maniksurtani/quotaservice/protos"
	"github.com/golang/protobuf/proto"
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

	client := qspb.NewQuotaServiceClient(conn)

	req := &qspb.AllowRequest{
		Namespace: proto.String("test.namespace"),
		Name: proto.String("abc"),
		NumTokensRequested: proto.Int64(1)}
	rsp, err := client.Allow(context.TODO(), req)
	if err != nil {
		fmt.Printf("Caught error %v", err)
	} else {
		fmt.Printf("Got response %v", rsp)
	}
}

