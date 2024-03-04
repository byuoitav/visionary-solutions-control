package device

import (
	"encoding/json"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/byuoitav/common/status"
	"github.com/byuoitav/common/structs"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type DeviceManager struct {
	Log      *zap.Logger
	Lvl      *zap.AtomicLevel
	ReqQueue chan VSRequest
}

func (dm *DeviceManager) RunHTTPServer(router *gin.Engine, port string) error {
	dm.Log.Info("registering http endpoints")

	dev := router.Group("")
	dev.GET("/input/:encoder/:decoder", dm.SetStreamHostHandler)
	dev.POST("/:address/videowall", dm.SetVideoWallHandler)
	dev.GET("input/get/:address", dm.GetConnectedHostHandler)
	dev.GET("/:address/hardware", dm.GetDeviceInfoHandler)
	dev.GET("/:address/signal", dm.GetStreamSignalHandler)
	dev.PUT("/configure/:encoder", dm.ConfigureDeviceHandler)

	server := http.Server{
		Addr:           port,
		MaxHeaderBytes: 1024 * 10,
	}

	return router.Run(server.Addr)
}

func (dm *DeviceManager) SetStreamHostHandler(c *gin.Context) {
	dm.Log.Debug("setting stream host")

	encoder := c.Param("encoder")
	decoder := c.Param("decoder")

	eIP := resolveIPAddress(encoder)
	dIP := resolveIPAddress(decoder)

	// check to see if the encoder is up?

	cmdStr := getCommandString(SWITCH_HOST)
	tokens := strings.Split(cmdStr, "temp")
	cmdStr = tokens[0] + eIP.IP.String() + tokens[1]

	respChan := make(chan VSResponse)
	defer close(respChan)

	req := VSRequest{
		Address:     dIP.IP.String(),
		Command:     cmdStr,
		RespChannel: respChan,
	}

	dm.ReqQueue <- req

	resp := <-respChan
	if resp.Error != nil {
		dm.Log.Error("failed to make request for setting stream host", zap.Error(resp.Error))
		c.JSON(http.StatusInternalServerError, "failed to make request to decoder")
		return
	}

	dm.Log.Debug("set stream host successfully", zap.String("encoder", encoder), zap.String("decoder", decoder))
	c.JSON(http.StatusOK, status.Input{Input: encoder})
}

func (dm *DeviceManager) SetVideoWallHandler(c *gin.Context) {
	dm.Log.Debug("setting video wall parameters")

	address := c.Param("address")
	var wallParams videoWallParams
	err := c.Bind(&wallParams)
	if err != nil {
		dm.Log.Error("failed to bind request body for setting video wall parameters", zap.Error(err))
		c.JSON(http.StatusBadRequest, "invalid request body")
		return
	}

	ip := resolveIPAddress(address)

	cmdStr := getCommandString(VIDEO_WALL)
	tokens := strings.Split(cmdStr, "temp")
	cmdStr = tokens[0] + strconv.Itoa(wallParams.TotalRows) + tokens[1] + strconv.Itoa(wallParams.TotalColumns) + tokens[2] + strconv.Itoa(wallParams.RowPosition) + tokens[3] + strconv.Itoa(wallParams.ColumnPosition) + tokens[4]

	respChan := make(chan VSResponse)
	defer close(respChan)

	req := VSRequest{
		Address:     ip.IP.String(),
		Command:     cmdStr,
		RespChannel: respChan,
	}

	dm.ReqQueue <- req

	resp := <-respChan
	if resp.Error != nil {
		dm.Log.Error("failed to make request for setting video wall parameters", zap.Error(resp.Error))
		c.JSON(http.StatusInternalServerError, "failed to make request to decoder")
		return
	}

	dm.Log.Debug("set video wall parameters successfully", zap.String("decoder", address))
	c.JSON(http.StatusOK, "ok")
}

func (dm *DeviceManager) GetConnectedHostHandler(c *gin.Context) {
	dm.Log.Debug("getting the current stream host")
	address := c.Param("address")

	ip := resolveIPAddress(address)

	cmdStr := getCommandString(GET_HOST)

	respChan := make(chan VSResponse)
	defer close(respChan)

	req := VSRequest{
		Address:     ip.IP.String(),
		Command:     cmdStr,
		RespChannel: respChan,
	}

	dm.ReqQueue <- req

	resp := <-respChan
	if resp.Error != nil {
		dm.Log.Error("failed to make request for getting device details", zap.Error(resp.Error))
		c.JSON(http.StatusInternalServerError, "failed to make request to decoder")
		return
	}

	streamHost := resp.Response["STREAM.HOST"]

	dm.Log.Debug("found the current stream host", zap.String("decoder", address), zap.String("encoder", streamHost))
	c.JSON(http.StatusOK, status.Input{Input: streamHost})
}

func (dm *DeviceManager) GetDeviceInfoHandler(c *gin.Context) {
	dm.Log.Debug("getting device details")

	address := c.Param("address")

	ip := resolveIPAddress(address)

	cmdStr := getCommandString(GET_INFO)

	respChan := make(chan VSResponse)
	defer close(respChan)

	req := VSRequest{
		Address:     ip.IP.String(),
		Command:     cmdStr,
		RespChannel: respChan,
	}

	dm.ReqQueue <- req

	resp := <-respChan
	if resp.Error != nil {
		dm.Log.Error("failed to make request for getting device details", zap.Error(resp.Error))
		c.JSON(http.StatusInternalServerError, "failed to make request to decoder")
		return
	}

	var info structs.HardwareInfo

	info.ModelName = resp.Response["UNIT.MODEL"]
	info.FirmwareVersion = resp.Response["UNIT.FIRMWARE"]
	info.BuildDate = resp.Response["UNIT.FIRMWARE_DATE"]
	info.PowerStatus = "" // uptime; not found with vs devices

	info.NetworkInfo.IPAddress = resp.Response["IP.ADDRESS"]
	info.NetworkInfo.MACAddress = resp.Response["UNIT.MAC_ADDRESS"]

	details, _ := json.Marshal(info)
	dm.Log.Debug("got device details", zap.String("device info", string(details)))
	c.JSON(http.StatusOK, info)
}

func (dm *DeviceManager) GetStreamSignalHandler(c *gin.Context) {
	dm.Log.Debug("getting signal data")

	address := c.Param("address")

	ip := resolveIPAddress(address)

	cmdStr := getCommandString(GET_SIGNAL)

	respChan := make(chan VSResponse)
	defer close(respChan)

	req := VSRequest{
		Address:     ip.IP.String(),
		Command:     cmdStr,
		RespChannel: respChan,
	}

	dm.ReqQueue <- req

	resp := <-respChan
	if resp.Error != nil {
		dm.Log.Error("failed to make request for getting device details", zap.Error(resp.Error))
		c.JSON(http.StatusInternalServerError, "failed to make request to device")
		return
	}

	timing := resp.Response["VIDEO.TIMING"]
	var signalStatus structs.ActiveSignal

	if timing == "Not Available" {
		dm.Log.Debug("no signal", zap.String("address", address))
		signalStatus.Active = false
	} else {
		dm.Log.Debug("active signal", zap.String("address", address))
		signalStatus.Active = true
	}

	dm.Log.Debug("returning signal status", zap.Bool("status", signalStatus.Active))
	c.JSON(http.StatusOK, signalStatus)
}

func (dm *DeviceManager) ConfigureDeviceHandler(c *gin.Context) {
	dm.Log.Debug("configuring encoder")

	encoder := c.Param("encoder")

	ip := resolveIPAddress(encoder)

	dm.Log.Debug("needed for av-api to work")
	c.JSON(http.StatusOK, status.Input{Input: ip.IP.String()})
}

func resolveIPAddress(host string) *net.IPAddr {
	ipAddr, err := net.ResolveIPAddr("ip", host)
	if err != nil {
		return nil
	}

	ipAddr.IP = ipAddr.IP.To4()
	return ipAddr
}
