package caddyhttpdportalpreview

import (
	"net/http"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"

	"github.com/inverse-inc/packetfence/go/httpdportalpreview"
	"github.com/inverse-inc/packetfence/go/panichandler"
	"github.com/inverse-inc/packetfence/go/plugin/caddy2"
)

func init() {
	caddy.RegisterModule(HttpDispatcherHandler{})
	httpcaddyfile.RegisterHandlerDirective("httpdportalpreview", caddy2.ParseCaddyfile[HttpDispatcherHandler])
}

type HttpDispatcherHandler struct {
	caddy2.ModuleBase
	proxy *httpdportalpreview.Proxy
}

func (h HttpDispatcherHandler) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.httpdportalpreview",
		New: func() caddy.Module { return &HttpDispatcherHandler{} },
	}
}

func (h *HttpDispatcherHandler) Provision(ctx caddy.Context) error {
	h.proxy = httpdportalpreview.NewProxy(ctx)
	return nil
}

func (h *HttpDispatcherHandler) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	ctx := r.Context()
	defer panichandler.Http(ctx, w)
	// This will never call the next middleware so make sure its the only «acting» middleware on this service
	h.proxy.ServeHTTP(w, r)
	// TODO change me and wrap actions into something that handles server errors
	return nil
}
