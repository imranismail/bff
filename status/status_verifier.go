// Reference Implementation: https://github.com/google/martian/blob/v3.1.0/status/status_verifier.go
// Copyright 2015 Google Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package status

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/martian/v3"
	"github.com/google/martian/v3/parse"
)

const errFormat = "response(%s) status code verify failure: got %d, want %d"

// Verifier verifies the status codes of all responses.
type Verifier struct {
	statusCode int
}

type verifierJSON struct {
	StatusCode int                  `json:"statusCode"`
	Scope      []parse.ModifierType `json:"scope"`
}

func init() {
	parse.Register("status.Verifier", verifierFromJSON)
}

// NewVerifier returns a new status.Verifier for statusCode.
func NewVerifier(statusCode int) martian.ResponseModifier {
	return &Verifier{
		statusCode: statusCode,
	}
}

// ModifyResponse verifies that the status code for all requests
// matches statusCode.
func (v *Verifier) ModifyResponse(res *http.Response) error {
	if res.StatusCode != v.statusCode {
		return fmt.Errorf(errFormat, res.Request.URL, res.StatusCode, v.statusCode)
	}

	return nil
}

// verifierFromJSON builds a status.Verifier from JSON.
//
// Example JSON:
// {
//   "status.Verifier": {
//     "scope": ["response"],
//     "statusCode": 401
//   }
// }
func verifierFromJSON(b []byte) (*parse.Result, error) {
	msg := &verifierJSON{}
	if err := json.Unmarshal(b, msg); err != nil {
		return nil, err
	}

	return parse.NewResult(NewVerifier(msg.StatusCode), msg.Scope)
}
