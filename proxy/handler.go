package proxy

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/smtp"
	"net/url"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/logsquaredn/rototiller"
	"github.com/logsquaredn/rototiller/api"
	"github.com/logsquaredn/rototiller/pb"
	files "github.com/swaggo/files"
	swagger "github.com/swaggo/gin-swagger"
)

type Handler struct {
	Key      string
	SMTPURL  *url.URL
	SMTPAuth smtp.Auth
	From     string
}

func NewHandler(ctx context.Context, proxyAddr, smtpAddr, smtpFrom, key string) (http.Handler, error) {
	var (
		_              = rototiller.LoggerFrom(ctx)
		router         = gin.New()
		swaggerHandler = swagger.WrapHandler(files.Handler)
		tokenParser    = jwt.NewParser()
		h              = &Handler{key, nil, nil, smtpFrom}
	)

	router.Use(gin.Recovery())

	u, err := url.Parse(proxyAddr)
	if err != nil {
		return nil, err
	}

	apiReverseProxy := httputil.NewSingleHostReverseProxy(u)

	if smtpAddr != "" {
		h.SMTPURL, err = url.Parse(smtpAddr)
		if err != nil {
			return nil, err
		}

		var (
			smtpUsername = os.Getenv("ROTOTILLER_SMTP_USERNAME")
			smtpPassword = os.Getenv("ROTOTILLER_SMTP_PASSWORD")
		)
		if smtpUsername != "" && smtpPassword != "" {
			h.SMTPAuth = smtp.PlainAuth("", smtpUsername, smtpPassword, h.SMTPURL.Hostname())
			h.SMTPURL.User = url.UserPassword(smtpUsername, smtpPassword)
		}
	}

	router.GET("/healthz", func(ctx *gin.Context) {
		ctx.Data(http.StatusOK, "application/text", []byte("ok\n"))
	})
	router.GET("/readyz", func(ctx *gin.Context) {
		ctx.Data(http.StatusOK, "application/text", []byte("ok\n"))
	})

	swagger := router.Group("/swagger/v1")
	{
		swagger.GET("/*any", func(ctx *gin.Context) {
			if ctx.Param("any") == "/" || ctx.Param("any") == "" {
				ctx.Redirect(http.StatusFound, "/swagger/v1/index.html")
			} else {
				swaggerHandler(ctx)
			}
		})
	}

	apiKey := router.Group("/api/v1/api-key")
	{
		apiKey.POST("", h.createApiKey)
	}

	router.NoRoute(func(ctx *gin.Context) {
		rawToken := ctx.GetHeader("Authorization")
		if rawToken == "" {
			ctx.JSON(http.StatusUnauthorized, &pb.Error{
				Message: "API key required",
			})
			return
		}
		ctx.Request.Header.Del("Authorization")

		token, err := tokenParser.ParseWithClaims(rawToken, &jwt.RegisteredClaims{}, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}

			return []byte(key), nil
		})
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, pb.NewErr(err))
			return
		}

		if err = token.Claims.Valid(); err != nil {
			ctx.JSON(http.StatusForbidden, pb.NewErr(err))
			return
		}

		if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok && token.Valid {
			ctx.Request.Header.Set(api.NamespaceHeader, claims.Subject)
		} else {
			ctx.JSON(http.StatusForbidden, &pb.Error{
				Message: "API key invalid",
			})
			return
		}

		apiReverseProxy.ServeHTTP(ctx.Writer, ctx.Request)
	})

	return router, nil
}
