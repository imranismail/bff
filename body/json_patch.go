package body

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	jsonpatch "github.com/evanphx/json-patch/v5"
	"github.com/google/martian/v3/log"
	"github.com/google/martian/v3/parse"
)

func init() {
	parse.Register("body.JSONPatch", jsonPatchModifierFromJSON)
}

type jsonPatchModifierJSON struct {
	Scope []parse.ModifierType `json:"scope"`
	Patch jsonpatch.Patch      `json:"patch"`
}

// JSONPatchModifier let you change the name of the fields of the generated responses
type JSONPatchModifier struct {
	patch jsonpatch.Patch
}

// NewJSONPatchModifier constructs and returns a body.JSONPatchModifier.
func NewJSONPatchModifier(patch jsonpatch.Patch) *JSONPatchModifier {
	log.Debugf("body.JSONPatch.New")
	return &JSONPatchModifier{
		patch: patch,
	}
}

// ModifyRequest patches the request body.
func (m *JSONPatchModifier) ModifyRequest(req *http.Request) error {
	log.Debugf("body.JSONPatch.ModifyRequest: request: %s", req.URL)

	original, err := ioutil.ReadAll(req.Body)

	if err != nil {
		return err
	}

	err = req.Body.Close()

	if err != nil {
		return err
	}

	modified, err := m.patch.Apply(original)

	if err != nil {
		return err
	}

	req.ContentLength = int64(len(modified))
	req.Body = ioutil.NopCloser(bytes.NewReader(modified))

	return nil
}

// ModifyResponse patches the response body.
func (m *JSONPatchModifier) ModifyResponse(res *http.Response) error {
	log.Debugf("body.JSONPatch.ModifyResponse: request: %s", res.Request.URL)

	original, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return err
	}

	res.Body.Close()

	modified, err := m.patch.Apply(original)

	if err != nil {
		return err
	}

	res.ContentLength = int64(len(modified))
	res.Body = ioutil.NopCloser(bytes.NewReader(modified))

	return nil
}

func jsonPatchModifierFromJSON(b []byte) (*parse.Result, error) {
	msg := &jsonPatchModifierJSON{}
	if err := json.Unmarshal(b, msg); err != nil {
		return nil, err
	}

	mod := NewJSONPatchModifier(msg.Patch)
	return parse.NewResult(mod, msg.Scope)
}
