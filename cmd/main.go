package main

import (
	"net/http"

	"github.com/byuoitav/visionary-solutions-control/device"

	"github.com/gin-gonic/gin"
	"github.com/spf13/pflag"
)

func main() {
	var logLevel, port, username, password string
	pflag.StringVarP(&port, "port", "p", "8042", "Port on which to run the http server")
	pflag.StringVarP(&logLevel, "log", "l", "Info", "Initial log level")
	pflag.StringVarP(&username, "username", "", "", "Username to access decoders/encoders")
	pflag.StringVarP(&password, "password", "", "", "Password to access decoders/encoders")
	pflag.Parse()

	log, logLvl := buildLogger(logLevel)

	log.Info("initializing device control...")

	queue := make(chan device.VSRequest)
	requestManager := device.RequestManager{
		ReqQueue: queue,
		Log:      log,
		Creds: device.DeviceCredentials{
			Username: username,
			Password: password,
		},
	}

	deviceManager := device.DeviceManager{
		Log:      log,
		Lvl:      logLvl,
		ReqQueue: queue,
	}

	go requestManager.HandleRequests()

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

		deviceManager.Lvl.SetLevel(level)
		ctx.String(http.StatusOK, lvl)
	})

	router.GET("/log-level", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, deviceManager.Log.Level().String())
	})

	err := deviceManager.RunHTTPServer(router, ":"+port)
	if err != nil {
		deviceManager.Log.Panic("http server failed")
	}
}
