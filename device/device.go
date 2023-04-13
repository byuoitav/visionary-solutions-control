package device

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type DeviceManager struct {
	Log *zap.Logger
	Lvl *zap.AtomicLevel
}

func (dm *DeviceManager) RunHTTPServer(router *gin.Engine, port string) error {
	dm.Log.Info("registering http endpoints")

	dev := router.Group("")
	dev.GET("/input/:transmitter/:receiver") // set receiver to transmission channel
	dev.POST("/:address/videowall")
	dev.GET("input/get/:address")
	dev.GET("/:address/hardware")
	dev.GET("/:address/signal")
	dev.PUT("/configure/:transmitter")

	server := http.Server{
		Addr:           port,
		MaxHeaderBytes: 1024 * 10,
	}

	return router.Run(server.Addr)
}
