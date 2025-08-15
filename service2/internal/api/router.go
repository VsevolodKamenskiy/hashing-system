package api

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"service2/internal/mw"
)

func NewRouter(h *Handlers, log *logrus.Logger) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(mw.RequestID())
	r.Use(mw.HTTPLogger(log))
	r.Use(mw.Metrics())

	r.POST("/send", h.Send)
	r.GET("/check", h.Check)
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	return r
}
