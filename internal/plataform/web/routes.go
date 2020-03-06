package web

import (
	"github.com/b2wdigital/restQL-golang/internal/plataform/conf"
	"github.com/b2wdigital/restQL-golang/internal/plataform/logger"
	"github.com/valyala/fasthttp"
	"net/http"
)

func API(config conf.Config, log logger.Logger) fasthttp.RequestHandler {
	app := NewApp(config, log)
	restQl := NewRestQl(config, log)

	app.Handle(http.MethodPost, "/validate-query", restQl.validateQuery)

	return app.RequestHandler()
}

func Health(config conf.Config, log logger.Logger) fasthttp.RequestHandler {
	app := NewApp(config, log)
	check := NewCheck(config.Build())

	app.Handle(http.MethodGet, "/health", check.health)
	app.Handle(http.MethodGet, "/resource-status", check.resourceStatus)

	return app.RequestHandlerWithoutMiddlewares()
}

func Debug(config conf.Config, log logger.Logger) fasthttp.RequestHandler {
	app := NewApp(config, log)
	pprof := NewPprof()

	app.Handle(http.MethodGet, "/debug/pprof/goroutine", pprof.Index)
	app.Handle(http.MethodGet, "/debug/pprof/heap", pprof.Index)
	app.Handle(http.MethodGet, "/debug/pprof/threadcreate", pprof.Index)
	app.Handle(http.MethodGet, "/debug/pprof/block", pprof.Index)
	app.Handle(http.MethodGet, "/debug/pprof/mutex", pprof.Index)

	app.Handle(http.MethodGet, "/debug/pprof/profile", pprof.Profile)

	return app.RequestHandlerWithoutMiddlewares()
}
