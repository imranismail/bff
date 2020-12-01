package body

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/google/martian/v3/log"
	"github.com/google/martian/v3/parse"
	"github.com/imranismail/bff/jsonpatch"
)

func init() {
	parse.Register("body.JSONPatch", jsonPatchModifierFromJSON)
}

type jsonPatchModifierJSON struct {
	Scope                    []parse.ModifierType `json:"scope"`
	Patch                    jsonpatch.Patch      `json:"patch"`
	SupportNegativeIndices   bool                 `json:"supportNegativeIndices"`
	AccumulatedCopySizeLimit int64                `json:"accumulatedCopySizeLimit"`
	SkipMissingPathOnRemove  bool                 `json:"skipMissingPathOnRemove"`
	SkipMissingPathOnMove    bool                 `json:"skipMissingPathOnMove"`
	SkipMissingPathOnCopy    bool                 `json:"skipMissingPathOnCopy"`
	SkipMissingPathOnReplace bool                 `json:"skipMissingPathOnReplace"`
	EnsurePathExistsOnAdd    bool                 `json:"ensurePathExistsOnAdd"`
}

// JSONPatchModifier let you change the name of the fields of the generated responses
type JSONPatchModifier struct {
	patch   *jsonpatch.Patch
	options *jsonpatch.ApplyOptions
}

// NewJSONPatchModifier constructs and returns a body.JSONPatchModifier.
func NewJSONPatchModifier(patch *jsonpatch.Patch, options *jsonpatch.ApplyOptions) *JSONPatchModifier {
	log.Debugf("body.JSONPatch.New")

	return &JSONPatchModifier{
		patch:   patch,
		options: options,
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

	modified, err := m.patch.ApplyWithOptions(original, m.options)

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

	modified, err := m.patch.ApplyWithOptions(original, m.options)

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

	mod := NewJSONPatchModifier(&msg.Patch, &jsonpatch.ApplyOptions{
		SupportNegativeIndices:   msg.SupportNegativeIndices,
		AccumulatedCopySizeLimit: msg.AccumulatedCopySizeLimit,
		SkipMissingPathOnRemove:  msg.SkipMissingPathOnRemove,
		SkipMissingPathOnMove:    msg.SkipMissingPathOnMove,
		SkipMissingPathOnCopy:    msg.SkipMissingPathOnCopy,
		SkipMissingPathOnReplace: msg.SkipMissingPathOnReplace,
		EnsurePathExistsOnAdd:    msg.EnsurePathExistsOnAdd,
	})

	return parse.NewResult(mod, msg.Scope)
}
