package rlgroup

import (
	"github.com/Zanda256/rate-limiter-go/foundation/logger"
	"github.com/Zanda256/rate-limiter-go/foundation/web"
	"net/http"
)

type Config struct {
	Log *logger.Logger
}

func Routes(app *web.App, cfg Config) {
	const version = "v1"

	//authen := mid.Authenticate(cfg.Auth)
	//ruleAdmin := mid.Authorize(cfg.Auth, auth.RuleAdminOnly)

	//usrCore := user.NewCore(cfg.Log, userdb.NewStore(cfg.Log, cfg.DB))

	hdl := New(cfg.Log)
	app.Handle(http.MethodPost, version, "/limited", hdl.Limited)
	app.Handle(http.MethodPost, version, "/unlimited", hdl.UnLimited)
}
