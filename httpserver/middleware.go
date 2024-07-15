package httpserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/teng231/gotools/common"
	"github.com/teng231/gotools/v2/httpclient"
)

type Handler = gin.HandlerFunc
type FnCheckBasicAuth func(u, p string) error
type FnEnsureClient func(clientId string) (string, error)

const (
	StartTracingKey = "ts"
)

type Signature struct {
	Signature string
	ClientId  string
	Timestamp int
}

type MiddleEngine struct {
	*gin.Engine
}

func UsingJWTParser() Handler {
	return func(c *gin.Context) {

	}
}

// UsingSignature get in query first, if got nothing get in header
func UsingSignature(fnEnsureClient FnEnsureClient) Handler {
	return func(c *gin.Context) {
		sig := &Signature{ClientId: c.Query("client_id"), Signature: c.Query("signature")}
		sig.Timestamp, _ = strconv.Atoi(c.Query("timestamp"))

		if sig.ClientId == "" {
			sig = &Signature{ClientId: c.GetHeader("client_id"), Signature: c.GetHeader("signature")}
			sig.Timestamp, _ = strconv.Atoi(c.GetHeader("timestamp"))
		}

		if sig.ClientId == "" || sig.Signature == "" || sig.Timestamp == 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, &ErrorWrapResponse{
				Error: "invalid_missing_params_to_verify"})
			return
		}
		if sig.Timestamp+2*60 < int(time.Now().Unix()) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, &ErrorWrapResponse{
				Error: "timestamp_invalid", Step: 1})
			return
		}
		if sig.Timestamp > int(time.Now().Unix()) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, &ErrorWrapResponse{
				Error: "timestamp_invalid", Step: 2})
			return
		}
		secret, err := fnEnsureClient(sig.ClientId)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, &ErrorWrapResponse{
				Error: "invalid_client_id", Step: 1})
			return
		}

		if sigHash := common.Hash(sig.ClientId, secret, sig.Timestamp); sigHash != sig.Signature {
			c.AbortWithStatusJSON(http.StatusUnauthorized, &ErrorWrapResponse{
				Error: "invalid_client_id", Step: 2})
			return
		}
		c.Next()
	}
}

func UsingCORSMode(domainAllowed string) Handler {
	return cors.New(cors.Config{
		AllowOrigins: strings.Split(domainAllowed, ","),
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

func UsingTimeTracing() Handler {
	return func(c *gin.Context) {
		c.Set(StartTracingKey, time.Now())
	}
}

type CaptchaResponse struct {
	Success bool `json:"success"`
}

func UsingCaptcha(captchaServer, token string) Handler {
	return func(c *gin.Context) {
		url := fmt.Sprintf("https://www.google.com/recaptcha/api/siteverify?secret=%s&response=%s", captchaServer, token)
		resp, err := httpclient.Exec(url,
			httpclient.WithMethod("POST"),
			httpclient.WithHeader(map[string]string{"Content-Type": "application/json"}),
			httpclient.WithBody(nil))
		if resp.HttpCode > 299 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, &ErrorWrapResponse{
				Error: "fail_to_authorized_captcha", Step: 1})
			return
		}
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, &ErrorWrapResponse{
				Error: "fail_to_authorized_captcha", Step: 2})
			return
		}

		var captResp *CaptchaResponse
		if err = json.Unmarshal(resp.Body, &captResp); err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, &ErrorWrapResponse{
				Error: "fail_to_authorized_captcha", Step: 3})
			return
		}

		if !captResp.Success {
			c.AbortWithStatusJSON(http.StatusUnauthorized, &ErrorWrapResponse{
				Error: "fail_to_authorized_captcha", Step: 4})
			return
		}
		c.Next()
	}
}

func UsingBasicAuth(fnCheckBasicAuth FnCheckBasicAuth) Handler {
	return func(c *gin.Context) {
		user, password, has := c.Request.BasicAuth()

		if !has {
			c.AbortWithStatusJSON(http.StatusUnauthorized, &ErrorWrapResponse{
				Error: "fail_to_authorized_with_basic_auth",
				Step:  1,
			})
			return
		}

		if err := fnCheckBasicAuth(user, password); err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, &ErrorWrapResponse{
				Error: "fail_to_authorized_with_basic_auth",
				Step:  2,
				Trace: err.Error(),
			})
			return
		}
		c.Next()
	}
}
