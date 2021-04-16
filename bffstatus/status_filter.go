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

package bffstatus

import (
	"encoding/json"
	"net/http"

	"github.com/google/martian/v3"
	"github.com/google/martian/v3/filter"
	"github.com/google/martian/v3/log"
	"github.com/google/martian/v3/parse"
)

var noop = martian.Noop("status.Filter")

func init() {
	parse.Register("status.Filter", filterFromJSON)
}

// Filter runs modifiers if the response status code matches the specified status code.
type Filter struct {
	*filter.Filter
}

type filterJSON struct {
	StatusCode   int                  `json:"statusCode"`
	Modifier     json.RawMessage      `json:"modifier"`
	ElseModifier json.RawMessage      `json:"else"`
	Scope        []parse.ModifierType `json:"scope"`
}

// Example JSON configuration message:
// {
//   "statusCode": 401,
//   "scope": ["request", "response"],
//   "modifier": { ... }
//   "else": { ... }
// }
func filterFromJSON(b []byte) (*parse.Result, error) {
	msg := &filterJSON{}
	if err := json.Unmarshal(b, msg); err != nil {
		return nil, err
	}

	filter := NewFilter(msg.StatusCode)

	m, err := parse.FromJSON(msg.Modifier)
	if err != nil {
		return nil, err
	}

	filter.RequestWhenTrue(m.RequestModifier())
	filter.ResponseWhenTrue(m.ResponseModifier())

	if len(msg.ElseModifier) > 0 {
		em, err := parse.FromJSON(msg.ElseModifier)
		if err != nil {
			return nil, err
		}

		if em != nil {
			filter.RequestWhenFalse(em.RequestModifier())
			filter.ResponseWhenFalse(em.ResponseModifier())
		}
	}

	return parse.NewResult(filter, msg.Scope)
}

// NewFilter constructs a filter that applies the modifer when the
// response status code matches the statusCode.
func NewFilter(statusCode int) *Filter {
	log.Debugf("status.NewFilter: %d", statusCode)

	m := NewMatcher(statusCode)
	f := filter.New()
	f.SetRequestCondition(m)
	f.SetResponseCondition(m)
	return &Filter{f}
}

// Matcher is a conditional evaluator of response statud code to be used in
// filters that take conditionals.
type Matcher struct {
	statusCode int
}

// NewMatcher builds a new status code matcher.
func NewMatcher(statusCode int) *Matcher {
	return &Matcher{
		statusCode: statusCode,
	}
}

// MatchRequest always returns false since request will not have any status code.
func (m *Matcher) MatchRequest(req *http.Request) bool {
	return false
}

// MatchResponse retuns true if m.StatusCode matches res.StatusCode.
func (m *Matcher) MatchResponse(res *http.Response) bool {
	matched := m.matches(res.StatusCode)

	if matched {
		log.Debugf("status.Matcher.MatchResponse: matched: %s", m.statusCode)
	}

	return matched
}

func (m *Matcher) matches(statusCode int) bool {
	return statusCode == m.statusCode
}
