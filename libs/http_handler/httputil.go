package httphandler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Arthur-Conti/guh/config"
	errorhandler "github.com/Arthur-Conti/guh/libs/error_handler"
	"github.com/Arthur-Conti/guh/libs/log/logger"
)

type HttpHandler struct {
	client http.Client
}

func NewHttpHandler() * HttpHandler {
	return &HttpHandler{
		client: http.Client{},
	}
}

func (hh *HttpHandler) Request(method, url string, body, result any) error {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			config.Config.Logger.Errorf(logger.LogMessage{ApplicationPackage: "httphandler", Message: "Error marshalling body: %v\n", Vals: []any{err}})
			return errorhandler.Wrap(errorhandler.InternalServerError, "Error marshalling body", err)
		}
		bodyReader = bytes.NewBuffer(b)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		config.Config.Logger.Errorf(logger.LogMessage{ApplicationPackage: "httphandler", Message: "Error creating request: %v\n", Vals: []any{err}})
		return errorhandler.Wrap(errorhandler.InternalServerError, "Error creating request", err)
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := hh.client.Do(req)
	if err != nil {
		config.Config.Logger.Errorf(logger.LogMessage{ApplicationPackage: "httphandler", Message: "Error sending get message to '%v': %v\n", Vals: []any{url, err}})
		return errorhandler.Wrap(errorhandler.ErrorType(resp.Status), "Error sending get message to "+url, err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		config.Config.Logger.Errorf(logger.LogMessage{ApplicationPackage: "httphandler", Message: "HTTP %d: %s", Vals: []any{resp.StatusCode, string(b)}})
		return errorhandler.New(errorhandler.ErrorType(resp.Status), fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(b)))
	}

	if resp != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			config.Config.Logger.Errorf(logger.LogMessage{ApplicationPackage: "heephandler", Message: "Error deconding body: %v\n", Vals: []any{err}})
			return errorhandler.Wrap(errorhandler.ErrorType(resp.Status), "Error decoding body", err)
		}
	}
	return nil
}