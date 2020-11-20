package body

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/google/martian/v3"
	"github.com/google/martian/v3/log"
	"github.com/google/martian/v3/parse"
)

var httpClient = &http.Client{
	Timeout: time.Second * 30,
}

func init() {
	parse.Register("body.JSONResource", jsonResourceModifierFromJSON)
}

type jsonResourceModifierJSON struct {
	Scope       []parse.ModifierType `json:"scope"`
	ResourceURL string               `json:"url"`
	Method      string               `json:"method"`
	Behavior    string               `json:"behavior"`
	Group       string               `json:"group"`
	Modifier    json.RawMessage      `json:"modifier"`
}

// JSONResource WIP
type JSONResource struct {
	body     []byte
	behavior string
	group    string
}

// JSONResourceModifier let you change the name of the fields of the generated responses
type JSONResourceModifier struct {
	resourceURL string
	method      string
	behavior    string
	group       string
	reqmod      martian.RequestModifier
	resmod      martian.ResponseModifier
}

func validBehavior(behavior string) bool {
	return behavior == "replace" || behavior == "merge"
}

// NewJSONResource WIP
func NewJSONResource(body []byte, behavior string, group string) *JSONResource {
	return &JSONResource{
		body:     body,
		behavior: behavior,
		group:    group,
	}
}

// ModifyResponse WIP
func (resource *JSONResource) ModifyResponse(res *http.Response) error {
	if resource.group != "" {
		group := make(map[string]json.RawMessage)
		msg := make(json.RawMessage, 0)

		err := json.Unmarshal(resource.body, &msg)

		if err != nil {
			return err
		}

		group[resource.group] = msg

		resource.body, err = json.Marshal(group)

		if err != nil {
			return err
		}
	}

	res.Header.Set("Content-Type", "application/json")
	res.Header.Del("Content-Encoding")

	switch resource.behavior {
	case "merge":
		original, err := ioutil.ReadAll(res.Body)

		if err != nil {
			return err
		}

		res.Body.Close()

		modified, err := jsonpatch.MergePatch(original, resource.body)

		if err != nil {
			return err
		}

		res.ContentLength = int64(len(modified))
		res.Body = ioutil.NopCloser(bytes.NewReader(modified))

		break

	case "replace":
		res.ContentLength = int64(len(resource.body))
		res.Body = ioutil.NopCloser(bytes.NewReader(resource.body))

		break
	}

	return nil
}

// NewJSONResourceModifier constructs and returns a body.JSONDataSourceModifier.
func NewJSONResourceModifier(method string, resourceURL string, behavior string, group string) (*JSONResourceModifier, error) {
	if behavior == "" {
		behavior = "merge"
	}

	if method == "" {
		method = "GET"
	}

	if !validBehavior(behavior) {
		return nil, fmt.Errorf("body.JSONResource.New: invalid behavior %q", behavior)
	}

	log.Debugf("body.JSONResource.New: method(%s) url(%s) behavior(%s)", method, resourceURL, behavior)

	m := &JSONResourceModifier{
		resourceURL: resourceURL,
		method:      method,
		behavior:    behavior,
		group:       group,
	}

	return m, nil
}

// SetRequestModifier Sets a RequestModifier
func (m *JSONResourceModifier) SetRequestModifier(reqmod martian.RequestModifier) {
	m.reqmod = reqmod
}

// SetResponseModifier Sets a ResponseModifier
func (m *JSONResourceModifier) SetResponseModifier(resmod martian.ResponseModifier) {
	m.resmod = resmod
}

// FetchResource fetches the resource
func (m *JSONResourceModifier) FetchResource() (martian.ResponseModifier, error) {
	log.Debugf("body.JSONResource.FetchResource: method(%s) url(%s)", m.method, m.resourceURL)

	req, err := http.NewRequest(
		m.method,
		m.resourceURL,
		bytes.NewBuffer([]byte{}),
	)

	if err != nil {
		return nil, err
	}

	if m.reqmod != nil {
		err = m.reqmod.ModifyRequest(req)

		if err != nil {
			return nil, err
		}
	}

	req.Header.Set("Accept", "application/json")

	res, err := httpClient.Do(req)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if m.resmod != nil {
		err = m.resmod.ModifyResponse(res)

		if err != nil {
			return nil, err
		}
	}

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return nil, err
	}

	resource := NewJSONResource(
		body,
		m.behavior,
		m.group,
	)

	return resource, nil
}

// ModifyResponse patches the response body.
func (m *JSONResourceModifier) ModifyResponse(res *http.Response) error {
	log.Debugf("body.JSONResource.ModifyResponse: request: %s", res.Request.URL)

	resource, err := m.FetchResource()

	if err != nil {
		return err
	}

	return resource.ModifyResponse(res)
}

func jsonResourceModifierFromJSON(b []byte) (*parse.Result, error) {
	msg := &jsonResourceModifierJSON{}
	if err := json.Unmarshal(b, msg); err != nil {
		return nil, err
	}

	mod, err := NewJSONResourceModifier(msg.Method, msg.ResourceURL, msg.Behavior, msg.Group)

	if err != nil {
		return nil, err
	}

	if msg.Modifier != nil {
		r, err := parse.FromJSON(msg.Modifier)

		if err != nil {
			return nil, err
		}

		reqmod := r.RequestModifier()

		if reqmod != nil {
			mod.SetRequestModifier(reqmod)
		}

		resmod := r.ResponseModifier()

		if resmod != nil {
			mod.SetResponseModifier(resmod)
		}
	}

	return parse.NewResult(mod, msg.Scope)
}
