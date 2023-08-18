package device

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"go.uber.org/zap"
)

type VSRequest struct {
	Address     string
	Command     string
	RespChannel chan VSResponse
}

type VSResponse struct {
	Response map[string]string
	Error    error
}

type DeviceCredentials struct {
	Username string
	Password string
}

func (c DeviceCredentials) toEncodedString() string {
	return base64.StdEncoding.EncodeToString([]byte(c.Username + ":" + c.Password))
}

type RequestManager struct {
	ReqQueue chan VSRequest
	Creds    DeviceCredentials
	Log      *zap.Logger
}

func (rm *RequestManager) HandleRequests() {
	for {
		req := <-rm.ReqQueue

		req.RespChannel <- rm.sendRequest(req.Address, req.Command)
	}
}

func (rm *RequestManager) sendRequest(address, commands string) VSResponse {
	rm.Log.Debug("sending request", zap.String("address", address))

	url := "http://" + address + "/cgi-bin/wapi.cgi"

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader([]byte(commands)))
	if err != nil {
		rm.Log.Error("could not form request", zap.String("address", address), zap.Error(err))
		return VSResponse{
			Response: nil,
			Error:    fmt.Errorf("could not form http request"),
		}
	}

	req.Header.Add("Content-type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", "Basic "+rm.Creds.toEncodedString())

	client := http.Client{
		Timeout: time.Second * 10,
	}

	resp, err := client.Do(req)
	if err != nil {
		rm.Log.Error("failure sending http request", zap.String("address", address), zap.Error(err))
		return VSResponse{
			Response: nil,
			Error:    fmt.Errorf("failure while sending http request"),
		}
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		rm.Log.Error("")
		return VSResponse{
			Response: nil,
			Error:    fmt.Errorf(""),
		}
	}

	respMap := parseResponse(string(b))
	if status, ok := respMap["API.STATUS"]; ok {
		re := regexp.MustCompile("SUCCESS")
		if re.Match([]byte(status)) {
			return VSResponse{
				Response: respMap,
				Error:    nil,
			}
		}
	}

	rm.Log.Error("received error response", zap.String("response", string(b)))
	return VSResponse{
		Response: nil,
		Error:    fmt.Errorf("request failed"),
	}
}

func parseResponse(s string) map[string]string {
	pairs := strings.Split(s, "&")
	ret := make(map[string]string)

	for _, p := range pairs {
		if len(p) > 0 {
			vals := strings.Split(p, "=")
			if len(vals) == 2 {
				ret[vals[0]] = vals[1]
			}
		}
	}

	return ret
}
