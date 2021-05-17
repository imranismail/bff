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

// Package bffquerystring contains a valuesMapModifier to rewrite query strings in a request.
package bffquerystring

import (
	"encoding/json"
	"net/http"

	"github.com/google/martian/v3"
	"github.com/google/martian/v3/parse"
	"github.com/google/martian/v3/log"
)

func init() {
	parse.Register("querystring.ValuesMap", valuesMapModifierFromJSON)
}

type ValuesMap struct {
	Field string
	Mapping map[string]string
}

type valuesMapModifier struct {
	maps []ValuesMap
}

type valuesMapModifierJSON struct {
	Maps []ValuesMap `json:"maps"`
	Scope []parse.ModifierType `json:"scope"`
}

func (m *valuesMapModifier) ModifyRequest(req *http.Request) error {
	query := req.URL.Query()

	for _, vmap := range m.maps {
		if value := query.Get(vmap.Field); value != "" {
			if newValue, ok := vmap.Mapping[value]; ok {
				query.Set(vmap.Field, newValue)
			}
		}
	}

	req.URL.RawQuery = query.Encode()

	return nil
}

func NewValuesMapModifier(maps []ValuesMap) martian.RequestModifier {
	return &valuesMapModifier{
		maps: maps,
	}
}

func valuesMapModifierFromJSON(b []byte) (*parse.Result, error) {
	msg := &valuesMapModifierJSON{}

	if err := json.Unmarshal(b, msg); err != nil {
		return nil, err
	}
	log.Debugf("%v", msg)

	return parse.NewResult(NewValuesMapModifier(msg.Maps), msg.Scope)
}

