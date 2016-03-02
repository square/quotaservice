// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/maniksurtani/quotaservice/master/LICENSE

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

// Administrable defines something that can be administered via this package.
type Administrable interface {
	Metrics() metrics.Metrics
	Configs() *configs.ServiceConfig
	BucketContainer() *buckets.BucketContainer
}

// ServeAdminConsole serves up an admin console for an Administrable over a http server.
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
