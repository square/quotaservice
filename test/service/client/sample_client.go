/*
 *   Copyright 2016 Manik Surtani
 *
 *   Licensed under the Apache License, Version 2.0 (the "License");
 *   you may not use this file except in compliance with the License.
 *   You may obtain a copy of the License at
 *
 *       http://www.apache.org/licenses/LICENSE-2.0
 *
 *   Unless required by applicable law or agreed to in writing, software
 *   distributed under the License is distributed on an "AS IS" BASIS,
 *   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *   See the License for the specific language governing permissions and
 *   limitations under the License.
 */

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
		Name: proto.String("one"),
		NumTokensRequested: proto.Int64(1)}
	rsp, err := client.Allow(context.TODO(), req)
	if err != nil {
		fmt.Printf("Caught error %v", err)
	} else {
		fmt.Printf("Got response %v", rsp)
	}
}

