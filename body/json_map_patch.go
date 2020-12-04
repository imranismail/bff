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
	parse.Register("body.JSONMapPatch", jsonMapPatchModifierFromJSON)
}

type jsonMapPatchModifierJSON struct {
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

// JSONMapPatchModifier let you change the name of the fields of the generated responses
type JSONMapPatchModifier struct {
	patch   *jsonpatch.Patch
	options *jsonpatch.ApplyOptions
}

// NewJSONMapPatchModifier constructs and returns a body.JSONPatchModifier.
func NewJSONMapPatchModifier(patch *jsonpatch.Patch, options *jsonpatch.ApplyOptions) *JSONMapPatchModifier {
	log.Debugf("body.JSONMapPatch.New")

	return &JSONMapPatchModifier{
		patch:   patch,
		options: options,
	}
}

// ModifyRequest patches the request body.
func (m *JSONMapPatchModifier) ModifyRequest(req *http.Request) error {
	log.Debugf("body.JSONMapPatch.ModifyRequest: request: %s", req.URL)

	var original []json.RawMessage
	var modified []byte

	err := json.NewDecoder(req.Body).Decode(&original)

	if err != nil {
		return err
	}

	err = req.Body.Close()

	if err != nil {
		return err
	}

	for i, d := range original {
		original[i], err = m.patch.ApplyWithOptions(d, m.options)

		if err != nil {
			return err
		}
	}

	if err != nil {
		return err
	}

	modified, err = json.Marshal(original)

	if err != nil {
		return err
	}

	req.ContentLength = int64(len(modified))
	req.Body = ioutil.NopCloser(bytes.NewReader(modified))

	return nil
}

// ModifyResponse patches the response body.
func (m *JSONMapPatchModifier) ModifyResponse(res *http.Response) error {
	log.Debugf("body.JSONMapPatch.ModifyResponse: request: %s", res.Request.URL)

	var original []json.RawMessage
	var modified []byte

	err := json.NewDecoder(res.Body).Decode(&original)

	if err != nil {
		return err
	}

	err = res.Body.Close()

	if err != nil {
		return err
	}

	for i, d := range original {
		original[i], err = m.patch.ApplyWithOptions(d, m.options)

		if err != nil {
			return err
		}
	}

	if err != nil {
		return err
	}

	modified, err = json.Marshal(original)

	if err != nil {
		return err
	}

	res.ContentLength = int64(len(modified))
	res.Body = ioutil.NopCloser(bytes.NewReader(modified))

	return nil
}

func jsonMapPatchModifierFromJSON(b []byte) (*parse.Result, error) {
	msg := &jsonMapPatchModifierJSON{}
	if err := json.Unmarshal(b, msg); err != nil {
		return nil, err
	}

	mod := NewJSONMapPatchModifier(&msg.Patch, &jsonpatch.ApplyOptions{
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
