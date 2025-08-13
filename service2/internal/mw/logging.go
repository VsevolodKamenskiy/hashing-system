package mw

import (
	"time"

	formatters "github.com/fabienm/go-logrus-formatters"
	graylog "github.com/gemnasium/logrus-graylog-hook/v3"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func NewLogger(service string) *logrus.Logger {
	log := logrus.New()
	log.SetFormatter(formatters.NewGelf(service))
	hook := graylog.NewGraylogHook("graylog:12201", map[string]interface{}{})
	log.Hooks.Add(hook)
	return log
}

func HTTPLogger(log *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		entry := log.WithFields(logrus.Fields{
			"request_id": FromContext(c.Request.Context()),
			"method":     c.Request.Method,
			"path":       c.Request.URL.Path,
			"status":     c.Writer.Status(),
			"duration":   time.Since(start).String(),
			"client_ip":  c.ClientIP(),
		})
		if len(c.Errors) > 0 {
			entry.WithField("errors", c.Errors.String()).Error("http request end with errors")
		} else {
			entry.Info("http request end")
		}
	}
}
