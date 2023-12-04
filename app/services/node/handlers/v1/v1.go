// Package v1 contains the full set of handler functions and routes
// supported by the v1 web api.
package v1

import (
	"net/http"

	"github.com/warlck/palladium/app/services/node/handlers/v1/public"
	"github.com/warlck/palladium/foundation/blockchain/state"
	"github.com/warlck/palladium/foundation/nameservice"
	"github.com/warlck/palladium/foundation/web"
	"go.uber.org/zap"
)

const version = "v1"

// Config contains all the mandatory systems required by handlers.
type Config struct {
	Log   *zap.SugaredLogger
	State *state.State
	NS    *nameservice.NameService
}

// PublicRoutes binds all the version 1 public routes.
func PublicRoutes(app *web.App, cfg Config) {
	pbl := public.Handlers{
		Log:   cfg.Log,
		State: cfg.State,
		NS:    cfg.NS,
	}

	app.Handle(http.MethodPost, version, "/tx/submit", pbl.SubmitWalletTransaction)
}
