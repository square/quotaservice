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
// TODO(manik) Implement this package
package admin

import (
	"fmt"
	"net/http"
	"github.com/maniksurtani/quotaservice/logging"
)

type AdminServer struct {
	port int
}

func NewAdminServer(port int) *AdminServer {
	s := AdminServer{port: port}
	return &s
}

func (this *AdminServer) Start() {
	logging.Printf("Starting admin console on port %v", this.port)
	http.HandleFunc("/", handler)
	go http.ListenAndServe(fmt.Sprintf(":%v", this.port), nil)
}

func (this *AdminServer) Stop() {
	logging.Print("Stopping admin console")
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "A future admin console")
}

