package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Serve struct {
	route *gin.Engine
}

func newServer() *Serve {
	r := &Serve{}
	r.route = gin.New()
	r.route.SetTrustedProxies(nil)
	r.route.Use(gin.Recovery(), gin.Logger(), enableCORsCheck())
	return r
}

func (r *Serve) apiMapping() *Serve {

	r.route.GET("/", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "[4client]👉Hope for the best, prepare for the worst")
	})

	r.route.GET("/ping", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	return r
}
