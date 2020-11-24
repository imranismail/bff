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

package bffurl

import (
	"net/http"
	"net/url"

	"github.com/google/martian/v3/log"
)

// Matcher is a conditional evaluator of request urls to be used in
// filters that take conditionals.
type Matcher struct {
	url     *url.URL
	pattern *Pattern
}

// NewMatcher builds a new url matcher.
func NewMatcher(url *url.URL) *Matcher {
	return &Matcher{
		url:     url,
		pattern: NewPattern(url.Path),
	}
}

// MatchRequest retuns true if all non-empty URL segments in m.url match the
// request URL.
func (m *Matcher) MatchRequest(req *http.Request) bool {
	matched := m.matches(req)

	if matched {
		log.Debugf("bffurl.Matcher.MatchRequest: matched: %s", req.URL)
	}

	return matched
}

// MatchResponse retuns true if all non-empty URL segments in m.url match the
// request URL.
func (m *Matcher) MatchResponse(res *http.Response) bool {
	matched := m.matches(res.Request)

	if matched {
		log.Debugf("bffurl.Matcher.MatchResponse: matched: %s", res.Request.URL)
	}

	return matched
}

// matches forces all non-empty URL segments to match or it returns false.
func (m *Matcher) matches(r *http.Request) bool {
	switch {
	case m.url.Scheme != "" && m.url.Scheme != r.URL.Scheme:
		return false
	case m.url.Host != "" && !MatchHost(r.URL.Host, m.url.Host):
		return false
	case m.url.Path != "" && !m.pattern.Match(r):
		return false
	case m.url.RawQuery != "" && m.url.RawQuery != r.URL.RawQuery:
		return false
	case m.url.Fragment != "" && m.url.Fragment != r.URL.Fragment:
		return false
	}

	return true
}
