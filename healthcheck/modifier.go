package healthcheck

import (
	"encoding/json"
	"net/http"

	"github.com/google/martian/v3"
	"github.com/google/martian/v3/log"
	"github.com/google/martian/v3/parse"
)

// Modifier alters the request method
type Modifier struct {
	statusCode int
}

type modifierJSON struct {
	StatusCode int                  `json:"statusCode"`
	Scope      []parse.ModifierType `json:"scope"`
}

func init() {
	parse.Register("bff.Healthcheck", modifierFromJSON)
}

func (m *Modifier) Match(req *http.Request) bool {
	return req.URL.Path == "/healthz" && req.Method == "GET"
}

func (m *Modifier) ModifyRequest(req *http.Request) error {
	if m.Match(req) {
		log.Debugf("bff.Healthcheck.ModifyRequest: %s", req.URL.String())
		ctx := martian.NewContext(req)
		ctx.SkipRoundTrip()
		ctx.SkipLogging()
	}
	return nil
}

// ModifyRequest sets the fields of req.URL to m.Url if they are not the zero value.
func (m *Modifier) ModifyResponse(res *http.Response) error {
	if m.Match(res.Request) {
		log.Debugf("bff.Healthcheck.ModifyResponse: %s", res.Request.URL.String())
		res.StatusCode = m.statusCode
	}
	return nil
}

// NewModifier overrides the url of the request.
func NewModifier(statusCode int) martian.RequestResponseModifier {
	log.Debugf("bff.Healthcheck.New: statusCode=%v", statusCode)

	if statusCode == 0 {
		statusCode = 200
	}

	return &Modifier{
		statusCode: statusCode,
	}
}

// modifierFromJSON builds a bffurl.Modifier from JSON.
//
// Example modifier JSON:
// {
//   "bff.MethodModifier": {
//     "scope": ["request"],
//     "method": "POST"
//   }
// }
func modifierFromJSON(b []byte) (*parse.Result, error) {
	msg := &modifierJSON{}
	if err := json.Unmarshal(b, msg); err != nil {
		return nil, err
	}

	mod := NewModifier(msg.StatusCode)

	return parse.NewResult(mod, msg.Scope)
}
