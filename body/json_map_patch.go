package body

import (
	"bytes"
	"encoding/json"
	"fmt"
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
	Path                     string               `json:"path"`
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
	path    string
}

// NewJSONMapPatchModifier constructs and returns a body.JSONPatchModifier.
func NewJSONMapPatchModifier(patch *jsonpatch.Patch, options *jsonpatch.ApplyOptions, path string) *JSONMapPatchModifier {
	log.Debugf("body.JSONMapPatch.New")

	return &JSONMapPatchModifier{
		patch:   patch,
		options: options,
		path:    path,
	}
}

func jsonMapPatch(original []byte, path string, patch *jsonpatch.Patch, options *jsonpatch.ApplyOptions) ([]byte, error) {
	var modified []byte

	doc, err := jsonpatch.FindObject(original, path)
	if err != nil {
		return nil, err
	}

	var array []json.RawMessage

	err = json.Unmarshal(doc, &array)
	if err != nil {
		return nil, err
	}

	for i, d := range array {
		array[i], err = patch.ApplyWithOptions(d, options)

		if err != nil {
			return nil, err
		}
	}

	modified, err = json.Marshal(array)
	if err != nil {
		return nil, err
	}

	if path != "" && path != "/" {
		p := fmt.Sprintf(`[{"op": "replace", "path": "%s", "value": [%s]}]`, path, modified)

		patch, err := jsonpatch.DecodePatch([]byte(p))
		if err != nil {
			return nil, err
		}

		modified, err = patch.Apply(original)
		if err != nil {
			return nil, err
		}
	}

	return modified, nil
}

// ModifyRequest patches the request body.
func (m *JSONMapPatchModifier) ModifyRequest(req *http.Request) error {
	log.Debugf("body.JSONMapPatch.ModifyRequest: request: %s", req.URL)

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return err
	}

	modified, err := jsonMapPatch(body, m.path, m.patch, m.options)
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

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	modified, err := jsonMapPatch(body, m.path, m.patch, m.options)
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
	}, msg.Path)

	return parse.NewResult(mod, msg.Scope)
}
