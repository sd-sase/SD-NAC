package pfipset

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/inverse-inc/go-utils/log"
	"github.com/inverse-inc/go-utils/sharedutils"
	"github.com/inverse-inc/packetfence/go/db"
	"github.com/inverse-inc/packetfence/go/panichandler"
	"github.com/inverse-inc/packetfence/go/pfconfigdriver"
	"github.com/inverse-inc/packetfence/go/plugin/caddy2"
)

// Queue value
const (
	maxQueueSize = 1000
	maxWorkers   = 1
)

// Register the plugin in caddy
func init() {
	caddy.RegisterModule(PfipsetHandler{})
	httpcaddyfile.RegisterHandlerDirective("pfipset", caddy2.ParseCaddyfile[PfipsetHandler])
}

type PfipsetHandler struct {
	caddy2.ModuleBase
	IPSET    *pfIPSET
	database *sql.DB
	router   *mux.Router
}

func (h PfipsetHandler) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.pfipset",
		New: func() caddy.Module { return &PfipsetHandler{} },
	}
}

func (h *PfipsetHandler) Provision(ctx caddy.Context) error {

	h.IPSET = &pfIPSET{}

	// create job channel
	h.IPSET.jobs = make(chan job, maxQueueSize)

	// create workers
	for i := 1; i <= maxWorkers; i++ {
		go func(i int) {
			for j := range h.IPSET.jobs {
				doWork(i, j)
			}
		}(i)
	}

	go func() {
		ctx := log.LoggerNewContext(context.Background())
		for {
			time.Sleep(1 * time.Second)
			currentQueueSize := len(h.IPSET.jobs)
			// Log a warning when queue is halfway full, and an error when its full
			if currentQueueSize >= maxQueueSize {
				log.LoggerWContext(ctx).Error("Queue has reached its maximum. Ipset related calls will be delayed and may timeout. Investigate previous logs to determine the cause of this backlog.")
			} else if currentQueueSize > (maxQueueSize * 0.5) {
				log.LoggerWContext(ctx).Warn(fmt.Sprintf("Queue has reached %d. Until it reaches %d, everything will still work.", currentQueueSize, maxQueueSize))
			}
		}
	}()

	// Default http timeout
	http.DefaultClient.Timeout = 10 * time.Second

	h.IPSET.detectType(ctx)

	db, err := db.DbFromConfig(ctx)
	sharedutils.CheckError(err)
	h.database = db

	// Reload the set from the database each 5 minutes
	// TODO: have this time configurable
	go func() {
		// Create a new context just for this goroutine
		ctx := context.WithValue(ctx, "dummy", "dummy")
		for {
			var Inline bool
			Inline = false
			var keyConfNet pfconfigdriver.PfconfigKeys
			keyConfNet.PfconfigNS = "config::Network"
			keyConfNet.PfconfigHostnameOverlay = "yes"

			pfconfigdriver.FetchDecodeSocket(ctx, &keyConfNet)

			for _, key := range keyConfNet.Keys {
				var ConfNet pfconfigdriver.RessourseNetworkConf
				ConfNet.PfconfigHashNS = key

				pfconfigdriver.FetchDecodeSocket(ctx, &ConfNet)

				if ConfNet.Type == "inlinel2" || ConfNet.Type == "inlinel3" || ConfNet.Type == "inline" {
					Inline = true
				}
			}
			if Inline {
				h.IPSET.initIPSet(ctx, h.database)
				log.LoggerWContext(ctx).Info("Reloading ipsets")
			} else {
				log.LoggerWContext(ctx).Info("No Inline Network bypass ipsets reload")
			}
			time.Sleep(300 * time.Second)
		}
	}()

	h.router = mux.NewRouter()
	api := h.router.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("/ipset/mark_layer3", handleLayer3).Methods("POST")
	api.HandleFunc("/ipset/mark_layer2", handleLayer2).Methods("POST")
	api.HandleFunc("/ipset/unmark_mac", h.IPSET.handleUnmarkMac).Methods("POST")
	api.HandleFunc("/ipset/unmark_ip", h.IPSET.handleUnmarkIp).Methods("POST")
	api.HandleFunc("/ipset/mark_ip_layer2", handleMarkIpL2).Methods("POST")
	api.HandleFunc("/ipset/mark_ip_layer3", handleMarkIpL3).Methods("POST")
	api.HandleFunc("/ipset/passthrough", handlePassthrough).Methods("POST")
	api.HandleFunc("/ipset/passthrough_isolation", handleIsolationPassthrough).Methods("POST")
	api.HandleFunc("/ipset/add_ip/{set_name}", handleAddIp).Methods("POST")
	api.HandleFunc("/ipset/remove_ip/{set_name}", handleRemoveIp).Methods("POST")

	return nil
}

func (h *PfipsetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	ctx := r.Context()
	ctx = h.IPSET.AddToContext(ctx)
	r = r.WithContext(ctx)

	defer panichandler.Http(ctx, w)

	routeMatch := mux.RouteMatch{}
	if h.router.Match(r, &routeMatch) {
		h.router.ServeHTTP(w, r)

		// TODO change me and wrap actions into something that handles server errors
		return nil
	}

	return next.ServeHTTP(w, r)
}
