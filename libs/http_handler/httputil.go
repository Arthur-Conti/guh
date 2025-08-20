package httphandler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	errorhandler "github.com/Arthur-Conti/guh/libs/error_handler"
	fl "github.com/Arthur-Conti/guh/libs/fast_logger"
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
			return errorhandler.Wrap(errorhandler.KindInternal, "Error marshalling body", err, errorhandler.WithOp("httphandler.Request"), errorhandler.WithFields(map[string]any{"method": method, "url": url, "body": body}))
		}
		bodyReader = bytes.NewBuffer(b)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return errorhandler.Wrap(errorhandler.KindInternal, "Error creating request", err, errorhandler.WithOp("httphandler.Request"), errorhandler.WithFields(map[string]any{"method": method, "url": url, "body": body}))
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := hh.client.Do(req)
	if err != nil {
		return errorhandler.Wrap(errorhandler.KindInternal, "Error sending get message to "+url, err, errorhandler.WithOp("httphandler.Request"), errorhandler.WithFields(map[string]any{"method": method, "url": url, "body": body}))
	}
	defer resp.Body.Close()
	
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return errorhandler.New(errorhandler.KindInternal, fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(b)), errorhandler.WithOp("httphandler.Request"), errorhandler.WithFields(map[string]any{"method": method, "url": url, "body": body}))
	}
	fl.Logf("Response: %+v", resp)
	if resp.Body != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return errorhandler.Wrap(errorhandler.KindInternal, "Error decoding body", err, errorhandler.WithOp("httphandler.Request"), errorhandler.WithFields(map[string]any{"method": method, "url": url, "body": body}))
		}
	}
	return nil
}