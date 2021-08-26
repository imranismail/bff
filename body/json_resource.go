package body

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/google/martian/v3"
	"github.com/google/martian/v3/log"
	"github.com/google/martian/v3/parse"
	"github.com/google/martian/v3/verify"
	"github.com/imranismail/bff/jsonpatch"
)

var httpClient = &http.Client{
	Timeout: time.Second * 30,
}

func init() {
	parse.Register("body.JSONResource", jsonResourceFromJSON)
}

type jsonResourceJSON struct {
	Scope          []parse.ModifierType `json:"scope"`
	ResourceURL    string               `json:"url"`
	Method         string               `json:"method"`
	Behavior       string               `json:"behavior"`
	Group          string               `json:"group"`
	AllowedHeaders []string             `json:"allowedHeaders"`
	Modifier       json.RawMessage      `json:"modifier"`
}

type jsonResource struct {
	body     []byte
	behavior string
	group    string
}

// JSONResource let you change the name of the fields of the generated responses
type JSONResource struct {
	resourceURL    string
	method         string
	behavior       string
	group          string
	allowedHeaders []string
	reqmod         martian.RequestModifier
	resmod         martian.ResponseModifier
}

func validBehavior(behavior string) bool {
	return behavior == "replace" || behavior == "merge"
}

// NewJSONResource constructs and returns a body.JSONDataSourceModifier.
func NewJSONResource(method string, resourceURL string, behavior string, group string, allowedHeaders []string) (*JSONResource, error) {
	if behavior == "" {
		behavior = "replace"
	}

	if method == "" {
		method = "GET"
	}

	if !validBehavior(behavior) {
		return nil, fmt.Errorf("body.JSONResource.New: invalid behavior %q", behavior)
	}

	log.Debugf("body.JSONResource.New: method(%s) url(%s) behavior(%s)", method, resourceURL, behavior)

	m := &JSONResource{
		resourceURL:    resourceURL,
		method:         method,
		behavior:       behavior,
		group:          group,
		allowedHeaders: allowedHeaders,
	}

	return m, nil
}

// SetRequestModifier Sets a RequestModifier
func (m *JSONResource) SetRequestModifier(reqmod martian.RequestModifier) {
	m.reqmod = reqmod
}

// SetResponseModifier Sets a ResponseModifier
func (m *JSONResource) SetResponseModifier(resmod martian.ResponseModifier) {
	m.resmod = resmod
}

// FetchResource fetches the resource
func (m *JSONResource) FetchResource(downstreamReq *http.Request) (martian.ResponseModifier, error) {
	log.Debugf("body.JSONResource.FetchResource: method(%s) url(%s) allowedHeaders(%s)", m.method, m.resourceURL, m.allowedHeaders)

	upstreamReq, err := http.NewRequest(
		m.method,
		m.resourceURL,
		bytes.NewBuffer([]byte{}),
	)

	if err != nil {
		return nil, err
	}

	upstreamReq.Header.Set("Accept", "application/json")

	for _, allowed := range m.allowedHeaders {
		header := downstreamReq.Header.Get(allowed)

		if header != "" {
			upstreamReq.Header.Add(allowed, header)
		}
	}

	_, cleanup, err := martian.TestContext(upstreamReq, nil, nil)

	if err != nil {
		return nil, err
	}

	defer cleanup()

	if m.reqmod != nil {
		err = m.reqmod.ModifyRequest(upstreamReq)

		if err != nil {
			return nil, err
		}
	}

	res, err := httpClient.Do(upstreamReq)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	res.Request = upstreamReq

	if err != nil {
		return nil, err
	}

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

	return &jsonResource{
		body:     body,
		behavior: m.behavior,
		group:    m.group,
	}, nil
}

// ModifyResponse patches the response body.
func (m *JSONResource) ModifyResponse(res *http.Response) error {
	log.Debugf("body.JSONResource.ModifyResponse: request: %s", res.Request.URL)

	resource, err := m.FetchResource(res.Request)

	if err != nil {
		return err
	}

	return resource.ModifyResponse(res)
}

// ResetResponseVerifications clears all failed response verifications.
func (m *JSONResource) ResetResponseVerifications() {
	if resv, ok := m.resmod.(verify.ResponseVerifier); ok {
		resv.ResetResponseVerifications()
	}
}

// VerifyResponses returns a MultiError containing all the
// verification errors returned by response verifiers.
func (m *JSONResource) VerifyResponses() error {
	log.Debugf("body.JSONResource.VerifyResponse")

	if resv, ok := m.resmod.(verify.ResponseVerifier); ok {
		if err := resv.VerifyResponses(); err != nil {
			return err
		}
	}

	return nil
}

func jsonResourceFromJSON(b []byte) (*parse.Result, error) {
	msg := &jsonResourceJSON{}

	if err := json.Unmarshal(b, msg); err != nil {
		return nil, err
	}

	m, err := NewJSONResource(msg.Method, msg.ResourceURL, msg.Behavior, msg.Group, msg.AllowedHeaders)

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
			m.SetRequestModifier(reqmod)
		}

		resmod := r.ResponseModifier()

		if resmod != nil {
			m.SetResponseModifier(resmod)
		}
	}

	return parse.NewResult(m, msg.Scope)
}

func (resource *jsonResource) ModifyResponse(res *http.Response) error {
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

	case "replace":
		res.ContentLength = int64(len(resource.body))
		res.Body = ioutil.NopCloser(bytes.NewReader(resource.body))
	}

	return nil
}
