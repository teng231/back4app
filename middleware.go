package main

import (
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func enableCORsCheck() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins: strings.Split(cfg.DomainAllowed, ","),
		AllowMethods: []string{"PUT", "GET", "POST", "DELETE", "OPTION"},
		AllowHeaders: []string{"Origin",
			"Access-Control-Allow-Origin",
			"Content-Type",
			"Content-Length",
			"Access-Control-Allow-Methods",
			"Accept-Encoding",
			"X-CSRF-Token",
			"Authorization",
			"x-captcha-token",
			"x-api-duration",
			"x-access-token",
			"token",
		},
		ExposeHeaders:    []string{"x-access-token", "x-api-duration", "x-captcha-token"},
		AllowWildcard:    true,
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}
