package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (r *Serve) handleHomelander(ctx *gin.Context) {
	ctx.String(http.StatusOK, "[4client]👉Hope for the best, prepare for the worst")
}

func (r *Serve) handlePing(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
}
