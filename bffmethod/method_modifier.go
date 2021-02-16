package bffmethod

import (
	"encoding/json"
	"net/http"

	"github.com/google/martian/v3"
	"github.com/google/martian/v3/log"
	"github.com/google/martian/v3/parse"
)

// Modifier alters the request method
type Modifier struct {
	method string
}

type modifierJSON struct {
	Method string               `json:"method"`
	Scope  []parse.ModifierType `json:"scope"`
}

func init() {
	parse.Register("bff.MethodModifier", modifierFromJSON)
}

// ModifyRequest sets the fields of req.URL to m.Url if they are not the zero value.
func (m *Modifier) ModifyRequest(req *http.Request) error {
	if m.method != "" {
		req.Method = m.method
	}

	return nil
}

// NewModifier overrides the url of the request.
func NewModifier(m string) martian.RequestModifier {
	log.Debugf("bff.NewMethodModifier: %s", m)

	return &Modifier{
		method: m,
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

	mod := NewModifier(msg.Method)

	return parse.NewResult(mod, msg.Scope)
}
