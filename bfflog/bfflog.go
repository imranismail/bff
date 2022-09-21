package bfflog

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/martian/v3"
	"github.com/google/martian/v3/parse"
	"github.com/imranismail/bff/log"
	"github.com/rs/zerolog"
)

// Logger is a modifier that logs requests and responses.
type Logger struct{}

type loggerJSON struct {
	Scope []parse.ModifierType `json:"scope"`
}

func init() {
	parse.Register("bfflog.Logger", loggerFromJSON)
}

// NewLogger returns a logger that logs requests and responses, optionally
// logging the body. Log function defaults to martian.Infof.
func NewLogger() *Logger {
	return &Logger{}
}

// ModifyRequest logs the request
func (l *Logger) ModifyRequest(req *http.Request) error {
	ctx := martian.NewContext(req)
	if ctx.SkippingLogging() {
		return nil
	}

	hdrs := zerolog.Dict()

	for key, val := range req.Header {
		hdrs.Str(key, strings.Join(val, ", "))
	}

	log.Logger.Zlog.Info().Str("path", req.URL.Path).Str("method", req.Method).Str("scheme", req.URL.Scheme).Str("host", req.Host).Dict("headers", hdrs).Msg(fmt.Sprintf("Request to %s", req.URL))

	return nil
}

// ModifyResponse logs the response
func (l *Logger) ModifyResponse(res *http.Response) error {
	ctx := martian.NewContext(res.Request)
	if ctx.SkippingLogging() {
		return nil
	}

	hdrs := zerolog.Dict()

	for key, val := range res.Header {
		hdrs.Str(key, strings.Join(val, ", "))
	}

	log.Logger.Zlog.Info().Str("path", res.Request.URL.Path).Str("method", res.Request.Method).Str("scheme", res.Request.URL.Scheme).Str("host", res.Request.Host).Dict("headers", hdrs).Msg(fmt.Sprintf("Response from %s", res.Request.URL))

	return nil
}

// loggerFromJSON builds a logger from JSON.
//
// Example JSON:
// {
//   "bfflog.Logger": {
//     "scope": ["request", "response"]
//   }
// }
func loggerFromJSON(b []byte) (*parse.Result, error) {
	msg := &loggerJSON{}
	if err := json.Unmarshal(b, msg); err != nil {
		return nil, err
	}

	l := NewLogger()

	return parse.NewResult(l, msg.Scope)
}
