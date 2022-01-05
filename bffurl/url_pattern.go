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

package bffurl

import (
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strings"

	"github.com/google/martian/v3"
)

var patternRe = regexp.MustCompile(`[/.;,]:([^/.;,]+)`)

// Pattern WIP
type Pattern struct {
	raw      string
	prefixes []string
	params   Params
	breaks   []byte
}

// Param WIP
type Param struct {
	name string
	idx  int
}

// RawName WIP
func (p *Param) RawName() string {
	return fmt.Sprintf(":%s", p.name)
}

// Name WIP
func (p *Param) Name() string {
	return `bffurl.ParamName.` + p.name
}

// Get WIP
func (p *Param) Get(ctx *martian.Context) string {
	if val, ok := ctx.Get(p.Name()); ok {
		return val.(string)
	}

	return ""
}

// Set WIP
func (p *Param) Set(req *http.Request, value string) {
	ctx := martian.NewContext(req)
	ctx.Set(p.Name(), value)
}

// Params WIP
type Params []Param

func (p Params) Len() int {
	return len(p)
}

func (p Params) Less(i, j int) bool {
	return p[i].name < p[j].name
}

func (p Params) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

// NewPattern WIP
func NewPattern(raw string) *Pattern {
	p := &Pattern{raw: raw}

	matches := patternRe.FindAllStringIndex(raw, -1)
	matchesLen := len(matches)

	p.params = make(Params, matchesLen)
	p.breaks = make([]byte, matchesLen)
	p.prefixes = make([]string, matchesLen+1)

	n := 0

	for i, match := range matches {
		start, end := match[0], match[1]

		p.prefixes[i] = raw[n : start+1]
		p.params[i].name = raw[start+2 : end]
		p.params[i].idx = i

		if end == len(raw) {
			p.breaks[i] = '/'
		} else {
			p.breaks[i] = raw[end]
		}

		n = end
	}

	p.prefixes[matchesLen] = raw[n:]

	sort.Sort(p.params)

	return p
}

// ReplaceParams WIP
func (p *Pattern) ReplaceParams(ctx *martian.Context, str string) string {
	for _, param := range p.params {
		str = strings.ReplaceAll(str, param.RawName(), param.Get(ctx))
	}

	return str
}

// Match WIP
func (p *Pattern) Match(r *http.Request) bool {
	path := r.URL.Path

	for i, param := range p.params {
		prefix := p.prefixes[i]

		if !strings.HasPrefix(path, prefix) {
			return false
		}

		path = path[len(prefix):]

		brk := p.breaks[i]
		n := 0

		for n < len(path) {
			if path[n] == brk || path[n] == '/' {
				break
			}

			n++
		}

		if n == 0 {
			return false
		}

		param.Set(r, path[:n])
		path = path[n:]
	}

	tail := p.prefixes[len(p.params)]

	return path == tail
}
