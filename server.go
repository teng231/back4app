package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/teng231/back4app/httpserver"
)

type Serve struct {
	*httpserver.Engine
}

func newServer() *Serve {
	r := &Serve{&httpserver.Engine{gin.New()}}
	r.SetTrustedProxies(nil)
	return r
}

func (r *Serve) customMiddleware() *Serve {
	r.Use(gin.Recovery(), gin.Logger())
	r.Use(httpserver.UsingCORSMode(cfg.DomainAllowed), httpserver.UsingTimeTracing())
	return r
}

func (r *Serve) registerHandlers() *Serve {

	r.GET("/", r.handleHomelander)
	r.GET("/ping", r.handlePing)

	r.GETs([]string{"/a", "/b"}, func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "[4client]👉Hope for the best, prepare for the worst")
	})

	r.GET("/response-err", httpserver.UsingTimeTracing(), func(ctx *gin.Context) {
		httpserver.JSON(ctx, http.StatusBadRequest, "[4client]👉Hope for the best, prepare for the worst")
	})

	return r
}
