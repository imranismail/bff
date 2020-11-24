package bffurl

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/google/martian/v3"
)

// Pattern WIP
type Pattern struct {
	raw      string
	prefixes []string
	params   []string
	breaks   []byte
}

var patternRe = regexp.MustCompile(`[/.;,]:([^/.;,]+)`)

// Name WIP
func Name(name string) string {
	return `bffurl.Name.` + name
}

// Param WIP
func Param(req *http.Request, name string) string {
	ctx := martian.NewContext(req)

	if val, ok := ctx.Get(Name(name)); ok {
		return val.(string)
	}

	return ""
}

// NewPattern WIP
func NewPattern(raw string) *Pattern {
	p := &Pattern{raw: raw}

	matches := patternRe.FindAllStringIndex(raw, -1)
	matchesLen := len(matches)

	p.params = make([]string, matchesLen)
	p.prefixes = make([]string, matchesLen)
	p.breaks = make([]byte, matchesLen)

	n := 0

	for i, match := range matches {
		start, end := match[0], match[1]

		p.prefixes[i] = raw[n : start+1]
		p.params[i] = raw[start+2 : end]

		if end == len(raw) {
			p.breaks[i] = '/'
		} else {
			p.breaks[i] = raw[end]
		}

		n = end
	}

	return p
}

// ReplacePath WIP
func (p *Pattern) ReplacePath(r *http.Request) {
	for _, param := range p.params {
		r.URL.Path = strings.ReplaceAll(r.URL.Path, fmt.Sprintf(":%s", param), Param(r, param))
	}
}

// Match WIP
func (p *Pattern) Match(r *http.Request) bool {
	ctx := martian.NewContext(r)
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

		ctx.Set(Name(param), path[:n])
		path = path[n:]
	}

	return true
}
