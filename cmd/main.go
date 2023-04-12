package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/spf13/pflag"
)

func main() {
	var logLevel, port string
	pflag.StringVarP(&port, "port", "p", "", "Port on which to run the http server")
	pflag.StringVarP(&logLevel, "log", "l", "Info", "Initial log level")
	pflag.Parse()

	log, logLvl := buildLogger(logLevel)

	log.Info("initializing device control...")

	router := gin.Default()

	router.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, "healthy")
	})

	router.GET("/status")

	router.PUT("/log-level/:level", func(ctx *gin.Context) {
		lvl := ctx.Param("level")

		level, err := getZapLevelFromString(lvl)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, "invalid log level")
			return
		}

		logLvl.SetLevel(level)
		ctx.String(http.StatusOK, lvl)
	})

	router.GET("/log-level", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, log.Level().String())
	})
}
