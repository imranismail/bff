package body

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/google/martian/v3"
	"github.com/google/martian/v3/log"
	"github.com/google/martian/v3/parse"
)

func init() {
	parse.Register("body.MultiFetcher", multiFetcherModifierFromJSON)
}

type multiFetcherJSON struct {
	Scope     []parse.ModifierType `json:"scope"`
	Resources []json.RawMessage    `json:"resources"`
}

// ResourceFetcher WIP
type ResourceFetcher interface {
	FetchResource() (martian.ResponseModifier, error)
}

// MultiFetcherModifier let you change the name of the fields of the generated responses
type MultiFetcherModifier struct {
	fetchers []ResourceFetcher
}

// NewMultiFetcherModifier constructs and returns a body.JSONDataSourceModifier.
func NewMultiFetcherModifier(fetchers []ResourceFetcher) *MultiFetcherModifier {
	return &MultiFetcherModifier{fetchers: fetchers}
}

// ModifyResponse patches the response body.
func (mod *MultiFetcherModifier) ModifyResponse(res *http.Response) error {
	log.Debugf("body.JSONDataSource.ModifyResponse: request: %s", res.Request.URL)

	resources := make(map[int]martian.ResponseModifier)
	mu := sync.Mutex{}
	merr := martian.NewMultiError()
	wg := sync.WaitGroup{}

	for i, fetcher := range mod.fetchers {
		wg.Add(1)

		go func(i int, fetcher ResourceFetcher) {
			defer wg.Done()

			resource, err := fetcher.FetchResource()

			if err != nil {
				mu.Lock()
				merr.Add(err)
				mu.Unlock()
			}

			mu.Lock()
			resources[i] = resource
			mu.Unlock()
		}(i, fetcher)
	}

	wg.Wait()

	mu.Lock()
	if !merr.Empty() {
		return merr
	}
	mu.Unlock()

	mu.Lock()
	for i := 0; i < len(resources); i++ {
		err := resources[i].ModifyResponse(res)

		if err != nil {
			return err
		}
	}
	mu.Unlock()

	return nil
}

func multiFetcherModifierFromJSON(b []byte) (*parse.Result, error) {
	msg := &multiFetcherJSON{}
	fetchers := make([]ResourceFetcher, 0)

	if err := json.Unmarshal(b, msg); err != nil {
		return nil, err
	}

	for _, resource := range msg.Resources {
		r, err := parse.FromJSON(resource)

		if err != nil {
			return nil, err
		}

		resmod := r.ResponseModifier()

		if fetcher, ok := resmod.(ResourceFetcher); ok {
			fetchers = append(fetchers, fetcher)
		}
	}

	mod := NewMultiFetcherModifier(fetchers)

	return parse.NewResult(mod, msg.Scope)
}
