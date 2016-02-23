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
	"github.com/maniksurtani/quotaservice/metrics"
	"github.com/maniksurtani/quotaservice/buckets"
	"github.com/maniksurtani/quotaservice/configs"
	"strings"
)

type Administrable interface {
	Metrics() metrics.Metrics
	Configs() *configs.ServiceConfig
	BucketContainer() *buckets.BucketContainer
}

func ServeAdminConsole(a Administrable, mux *http.ServeMux) {
	logging.Print("Serving admin console")
	mux.Handle("/", &handler{a})
}

type handler struct {
	a Administrable
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-type", "text/html")
	fmt.Fprintf(w, `
		<HTML>
			<BODY>
				<H1>A Future Admin Console</H1>
				For now, here's some information:
				<H3>Configuration</H3>
				%v
				<H3>Active buckets</H3>
				%v
			</BODY>
		</HTML>
	`, toHtml(h.a.Configs()), toHtml(h.a.BucketContainer()))
}

func toHtml(s interface{}) string {
	return fmt.Sprintf(`
<DIV><PRE>
%v
</PRE></DIV>
	`, strings.Replace(fmt.Sprintf("%v", s), "\n", "<BR />", -1))
}
