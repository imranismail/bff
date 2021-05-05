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

// Package bffquerystring contains a modifier to rewrite query strings in a request.
package bffquerystring

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"

	"github.com/google/martian/v3"
	"github.com/google/martian/v3/parse"
)

func init() {
	parse.Register("bff.QuerystringModifier", modifierFromJSON)
}

type modifier struct {
	op, key, value string
}

type modifierJSON struct {
	Key   string               `json:"name"`
	Value string               `json:"value"`
	Op    string               `json:"op"`
	Scope []parse.ModifierType `json:"scope"`
}


func replaceParams(value string, req *http.Request) string {
		re := regexp.MustCompile(`^:([0-9A-Za-z_]+)$`)
		matches := re.FindStringSubmatch(value)
		if len(matches) == 2 {
			param := "bffurl.ParamName." + matches[1]
			ctx := martian.NewContext(req)
			if val, ok := ctx.Get(param); ok {
				value = val.(string)
			}
		}
		return value
}

// ModifyRequest modifies the query string of the request with the given key and value.
func (m *modifier) ModifyRequest(req *http.Request) error {
	query := req.URL.Query()
	v := query.Get(m.key)

	switch m.op {
	case "add":
		if v == "" {
			value := replaceParams(m.value, req)
			query.Set(m.key, value)
		}

	case "replace":
		if v != "" {
			value := replaceParams(m.value, req)
			query.Set(m.key, value)
		}

	case "delete":
		if v != "" {
			query.Del(m.key)
		}

	case "copy":
		if v != "" {
			query.Set(m.value, v)
		}

	case "move":
		if v != "" {
			query.Set(m.value, v)
			query.Del(m.key)
		}

	default:
		return errors.New(fmt.Sprintf("bffquerystring.Modifier: Unknown operation '%s'", m.op))
	}
	req.URL.RawQuery = query.Encode()

	return nil
}

// NewModifier returns a request modifier that will set the query string
// at key with the given value. If the query string key already exists all
// values will be overwritten.
func NewModifier(op, key, value string) martian.RequestModifier {
	return &modifier{
		op:    op,
		key:   key,
		value: value,
	}
}

// modifierFromJSON takes a JSON message as a byte slice and returns
// a bffquerystring.modifier and an error.
//
// Example JSON:
// {
//  "name": "param",
//  "value": "true",
//  "scope": ["request", "response"]
// }
func modifierFromJSON(b []byte) (*parse.Result, error) {
	msg := &modifierJSON{}

	if err := json.Unmarshal(b, msg); err != nil {
		return nil, err
	}

	return parse.NewResult(NewModifier(msg.Op, msg.Key, msg.Value), msg.Scope)
}
