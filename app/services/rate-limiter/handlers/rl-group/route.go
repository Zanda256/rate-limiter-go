package rl_group

import (
	"github.com/Zanda256/rate-limiter-go/foundation/web"
	"net/http"
)

func Routes(app *web.App, cfg Config) {
	const version = "v1"

	//authen := mid.Authenticate(cfg.Auth)
	//ruleAdmin := mid.Authorize(cfg.Auth, auth.RuleAdminOnly)

	//usrCore := user.NewCore(cfg.Log, userdb.NewStore(cfg.Log, cfg.DB))

	hdl := New(usrCore, cfg.Auth)
	app.Handle(http.MethodPost, version, "/limited")
	app.Handle(http.MethodPost, version, "/unlimited")
}
