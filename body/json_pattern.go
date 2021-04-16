// Adapted from https://github.com/goji/goji/blob/v3.0.0/pat/pat.go
// Copyright (c) 2015, 2016 Carl Jackson (carl@avtok.com)

// MIT License

// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
// FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package body

import (
	"bytes"
	"fmt"
	"net/http"
	"regexp"

	"github.com/google/martian/v3"
)

type Pattern struct {
	params []Param
}

type Param string

func (p Param) label() string {
	return fmt.Sprintf(":%s", p)
}

func (p Param) name() string {
	return fmt.Sprintf("bffurl.ParamName.%s", p)
}

func (p Param) get(req *http.Request) string {
	ctx := martian.NewContext(req)

	if val, ok := ctx.Get(p.name()); ok {
		return val.(string)
	}

	return ""
}

func NewPattern(raw []byte) *Pattern {
	type void struct{}
	set := make(map[string]void)
	var v void

	re := regexp.MustCompile(`":([0-9A-Za-z_]+)"`)
	matches := re.FindAllSubmatch(raw, -1)

	for _, match := range matches {
		set[string(match[1])] = v
	}

	var params []Param

	for param := range set {
		params = append(params, Param(param))
	}

	p := Pattern{params}
	return &p
}

func (p *Pattern) ReplaceParams(body []byte, r *http.Request) []byte {
	for _, param := range p.params {
		val := param.get(r)
		if val == "" {
			val = param.label()
		}

		body = bytes.ReplaceAll(body, []byte(param.label()), []byte(val))
	}
	return body
}
