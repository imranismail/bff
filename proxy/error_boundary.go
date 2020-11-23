package proxy

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/google/martian/v3"
	"github.com/google/martian/v3/log"
	"github.com/google/martian/v3/verify"
)

// ErrorBoundary WIP
type ErrorBoundary struct {
	reqmod martian.RequestModifier
	resmod martian.ResponseModifier
	reqv   verify.RequestVerifier
	resv   verify.ResponseVerifier
}

type boundaryResponse struct {
	Errors []boundaryError `json:"errors"`
}

type boundaryError struct {
	Message string `json:"message"`
}

// NewErrorBoundary WIP
func NewErrorBoundary() *ErrorBoundary {
	return &ErrorBoundary{}
}

// SetRequestModifier WIP
func (eb *ErrorBoundary) SetRequestModifier(reqmod martian.RequestModifier) {
	eb.reqmod = reqmod
}

// SetResponseModifier WIP
func (eb *ErrorBoundary) SetResponseModifier(resmod martian.ResponseModifier) {
	eb.resmod = resmod
}

// SetRequestVerifier WIP
func (eb *ErrorBoundary) SetRequestVerifier(reqv verify.RequestVerifier) {
	eb.reqv = reqv
}

// SetResponseVerifier WIP
func (eb *ErrorBoundary) SetResponseVerifier(resv verify.ResponseVerifier) {
	eb.resv = resv
}

// ModifyRequest WIP
func (eb *ErrorBoundary) ModifyRequest(req *http.Request) error {
	defer eb.reqv.ResetRequestVerifications()

	merr := martian.NewMultiError()

	if err := eb.reqmod.ModifyRequest(req); err != nil {
		merr.Add(err)
	}

	if err := eb.reqv.VerifyRequests(); err != nil {
		merr.Add(err)
	}

	if !merr.Empty() {
		return merr
	}

	return nil
}

// ModifyResponse WIP
func (eb *ErrorBoundary) ModifyResponse(res *http.Response) error {
	defer eb.resv.ResetResponseVerifications()

	merr := martian.NewMultiError()

	if eb.resmod != nil {
		if err := eb.resmod.ModifyResponse(res); err != nil {
			merr.Add(err)
		}
	}

	if eb.resv != nil {
		if err := eb.resv.VerifyResponses(); err != nil {
			merr.Add(err)
		}
	}

	if !merr.Empty() {
		log.Errorf("proxy.ErrorBoundary.ModifyResponse: %v", merr)

		res.Body.Close()

		resp, err := merrToJSON(merr)

		if err != nil {
			return err
		}

		res.ContentLength = int64(len(resp))
		res.Body = ioutil.NopCloser(bytes.NewReader(resp))
	}

	return nil
}

func merrToJSON(merr *martian.MultiError) ([]byte, error) {
	errs := merr.Errors()

	vres := &boundaryResponse{
		Errors: make([]boundaryError, 0),
	}

	for _, err := range errs {
		vres.Errors = append(vres.Errors, boundaryError{Message: err.Error()})
	}

	return json.Marshal(vres)
}
