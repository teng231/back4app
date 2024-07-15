package httpserver

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type Engine struct {
	*gin.Engine
}

func HTTPStart(e *Engine, port string, wDuration, rDuration time.Duration) *http.Server {
	httpServer := &http.Server{
		Addr:        ":" + port,
		Handler:     e.Handler(),
		ReadTimeout: 3 * time.Minute, WriteTimeout: 75 * time.Second,
	}
	return httpServer
}
func JSON(ctx *gin.Context, code int, obj any) {
	if val, has := ctx.Get(StartTracingKey); has {
		s := time.Since(val.(time.Time))
		ctx.Header("x-api-duration", s.String())
	}

	if code >= 200 && code <= 299 {
		ctx.JSON(code, obj)
		return
	}

	if !strings.Contains(reflect.TypeOf(obj).String(), "ErrorWrapResponse") {
		ctx.JSON(code, &ErrorWrapResponse{
			Error: "error_invalid_format",
			Trace: fmt.Sprintf("warnning: error not trust. Format invalid. Path: %s", ctx.Request.URL.Path),
		})
		return
	}
	ctx.JSON(code, obj)
}

func (r *Engine) GETs(relativePaths []string, handlers ...gin.HandlerFunc) {
	for _, relativePath := range relativePaths {
		r.Handle(http.MethodGet, relativePath, handlers...)
	}
}

func (r *Engine) POSTs(relativePaths []string, handlers ...gin.HandlerFunc) {
	for _, relativePath := range relativePaths {
		r.Handle(http.MethodPost, relativePath, handlers...)
	}
}
